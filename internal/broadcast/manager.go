// Copyright © 2022 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broadcast

import (
	"context"

	"github.com/hyperledger/firefly/internal/batch"
	"github.com/hyperledger/firefly/internal/batchpin"
	"github.com/hyperledger/firefly/internal/config"
	"github.com/hyperledger/firefly/internal/data"
	"github.com/hyperledger/firefly/internal/i18n"
	"github.com/hyperledger/firefly/internal/identity"
	"github.com/hyperledger/firefly/internal/log"
	"github.com/hyperledger/firefly/internal/metrics"
	"github.com/hyperledger/firefly/internal/operations"
	"github.com/hyperledger/firefly/internal/syncasync"
	"github.com/hyperledger/firefly/internal/sysmessaging"
	"github.com/hyperledger/firefly/pkg/blockchain"
	"github.com/hyperledger/firefly/pkg/database"
	"github.com/hyperledger/firefly/pkg/dataexchange"
	"github.com/hyperledger/firefly/pkg/fftypes"
	"github.com/hyperledger/firefly/pkg/sharedstorage"
)

const broadcastDispatcherName = "pinned_broadcast"

type Manager interface {
	fftypes.Named

	NewBroadcast(ns string, in *fftypes.MessageInOut) sysmessaging.MessageSender
	BroadcastDatatype(ctx context.Context, ns string, datatype *fftypes.Datatype, waitConfirm bool) (msg *fftypes.Message, err error)
	BroadcastNamespace(ctx context.Context, ns *fftypes.Namespace, waitConfirm bool) (msg *fftypes.Message, err error)
	BroadcastMessage(ctx context.Context, ns string, in *fftypes.MessageInOut, waitConfirm bool) (out *fftypes.Message, err error)
	BroadcastDefinitionAsNode(ctx context.Context, ns string, def fftypes.Definition, tag string, waitConfirm bool) (msg *fftypes.Message, err error)
	BroadcastDefinition(ctx context.Context, ns string, def fftypes.Definition, signingIdentity *fftypes.SignerRef, tag string, waitConfirm bool) (msg *fftypes.Message, err error)
	BroadcastIdentityClaim(ctx context.Context, ns string, def *fftypes.IdentityClaim, signingIdentity *fftypes.SignerRef, tag string, waitConfirm bool) (msg *fftypes.Message, err error)
	BroadcastTokenPool(ctx context.Context, ns string, pool *fftypes.TokenPoolAnnouncement, waitConfirm bool) (msg *fftypes.Message, err error)
	Start() error
	WaitStop()

	// From operations.OperationHandler
	PrepareOperation(ctx context.Context, op *fftypes.Operation) (*fftypes.PreparedOperation, error)
	RunOperation(ctx context.Context, op *fftypes.PreparedOperation) (complete bool, err error)
}

type broadcastManager struct {
	ctx                   context.Context
	database              database.Plugin
	identity              identity.Manager
	data                  data.Manager
	blockchain            blockchain.Plugin
	exchange              dataexchange.Plugin
	sharedstorage         sharedstorage.Plugin
	batch                 batch.Manager
	syncasync             syncasync.Bridge
	batchpin              batchpin.Submitter
	maxBatchPayloadLength int64
	metrics               metrics.Manager
	operations            operations.Manager
}

func NewBroadcastManager(ctx context.Context, di database.Plugin, im identity.Manager, dm data.Manager, bi blockchain.Plugin, dx dataexchange.Plugin, si sharedstorage.Plugin, ba batch.Manager, sa syncasync.Bridge, bp batchpin.Submitter, mm metrics.Manager, om operations.Manager) (Manager, error) {
	if di == nil || im == nil || dm == nil || bi == nil || dx == nil || si == nil || ba == nil || mm == nil || om == nil {
		return nil, i18n.NewError(ctx, i18n.MsgInitializationNilDepError)
	}
	bm := &broadcastManager{
		ctx:                   ctx,
		database:              di,
		identity:              im,
		data:                  dm,
		blockchain:            bi,
		exchange:              dx,
		sharedstorage:         si,
		batch:                 ba,
		syncasync:             sa,
		batchpin:              bp,
		maxBatchPayloadLength: config.GetByteSize(config.BroadcastBatchPayloadLimit),
		metrics:               mm,
		operations:            om,
	}

	bo := batch.DispatcherOptions{
		RequiresSharedDataPayloadRefs: true,
		BatchMaxSize:                  config.GetUint(config.BroadcastBatchSize),
		BatchMaxBytes:                 bm.maxBatchPayloadLength,
		BatchTimeout:                  config.GetDuration(config.BroadcastBatchTimeout),
		DisposeTimeout:                config.GetDuration(config.BroadcastBatchAgentTimeout),
	}

	ba.RegisterDispatcher(broadcastDispatcherName,
		fftypes.TransactionTypeBatchPin,
		[]fftypes.MessageType{
			fftypes.MessageTypeBroadcast,
			fftypes.MessageTypeDefinition,
			fftypes.MessageTypeTransferBroadcast,
		}, bm.dispatchBatch, bo)

	om.RegisterHandler(ctx, bm, []fftypes.OpType{
		fftypes.OpTypeSharedStorageBatchBroadcast,
	})

	return bm, nil
}

func (bm *broadcastManager) Name() string {
	return "BroadcastManager"
}

func (bm *broadcastManager) dispatchBatch(ctx context.Context, batch *fftypes.Batch, pins []*fftypes.Bytes32) error {
	// The completed SharedStorage upload
	op := fftypes.NewOperation(
		bm.sharedstorage,
		batch.Namespace,
		batch.Payload.TX.ID,
		fftypes.OpTypeSharedStorageBatchBroadcast)
	addBatchBroadcastInputs(op, batch.ID)
	if err := bm.operations.AddOrReuseOperation(ctx, op); err != nil {
		return err
	}
	if err := bm.operations.RunOperation(ctx, opBatchBroadcast(op, batch)); err != nil {
		return err
	}
	log.L(ctx).Infof("Pinning broadcast batch %s with author=%s key=%s", batch.ID, batch.Author, batch.Key)
	return bm.batchpin.SubmitPinnedBatch(ctx, batch, pins)
}

func (bm *broadcastManager) publishBlobs(ctx context.Context, dataToPublish []*fftypes.DataAndBlob) error {
	for _, d := range dataToPublish {
		// Stream from the local data exchange ...
		reader, err := bm.exchange.DownloadBLOB(ctx, d.Blob.PayloadRef)
		if err != nil {
			return i18n.WrapError(ctx, err, i18n.MsgDownloadBlobFailed, d.Blob.PayloadRef)
		}
		defer reader.Close()

		// ... to the shared storage
		sharedRef, err := bm.sharedstorage.PublishData(ctx, reader)
		if err != nil {
			return err
		}
		log.L(ctx).Infof("Published blob with hash '%s' for data '%s' to shared storage: '%s'", d.Blob.Hash, d.Data.ID, sharedRef)

		// Update the data in the database, with the shared reference.
		// We do this independently for each piece of data
		update := database.DataQueryFactory.NewUpdate(ctx).Set("blob.public", sharedRef)
		err = bm.database.UpdateData(ctx, d.Data.ID, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bm *broadcastManager) Start() error {
	return nil
}

func (bm *broadcastManager) WaitStop() {
	// No go routines
}
