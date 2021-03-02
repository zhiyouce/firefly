package io.kaleido.kat.states;

import io.kaleido.kat.contracts.AssetTrailContract;
import net.corda.core.contracts.BelongsToContract;
import net.corda.core.identity.AbstractParty;
import net.corda.core.identity.Party;
import org.jetbrains.annotations.NotNull;

import java.util.ArrayList;
import java.util.List;

@BelongsToContract(AssetTrailContract.class)
public class AssetInstancePropertySet implements AssetEventState {
    private final String assetDefinitionID;
    private final String assetInstanceID;
    private final Party author;
    private final String key;
    private final String value;
    private final List<Party> participants;

    public AssetInstancePropertySet(String assetDefinitionID, String assetInstanceID, Party author, String key, String value, List<Party> participants) {
        this.assetDefinitionID = assetDefinitionID;
        this.assetInstanceID = assetInstanceID;
        this.author = author;
        this.key = key;
        this.value = value;
        this.participants = participants;
    }

    @NotNull
    @Override
    public List<AbstractParty> getParticipants() {
        return new ArrayList<>(participants);
    }

    @Override
    public String toString() {
        return String.format("AssetInstancePropertySet(assetDefinitionID=%s, assetInstanceID=%s, author=%s, key=%s, value=%s, participants=%s)", assetDefinitionID, assetInstanceID, author, key, value, participants);
    }

    @Override
    public Party getAuthor() {
        return author;
    }


    public String getAssetInstanceID() {
        return assetInstanceID;
    }

    public String getKey() {
        return key;
    }

    public String getValue() {
        return value;
    }

    public String getAssetDefinitionID() {
        return assetDefinitionID;
    }
}
