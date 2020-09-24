package com.samouraiwallet.soroban;

import com.codahale.xsalsa20poly1305.Keys;

import org.apache.commons.codec.DecoderException;
import org.apache.commons.codec.binary.Hex;

public class User {
    byte[] privateKey = Keys.generatePrivateKey();

    public String publicKey() {
        return Hex.encodeHexString(Keys.generatePublicKey(privateKey));
    }

    public Box box(String otherPublicKey) throws DecoderException {
        return new Box(privateKey, Hex.decodeHex(otherPublicKey));
    }

    public String sharedSecret(Box box) {
        byte[] sharedSecret = box.sharedSecret();
        return Hex.encodeHexString(sharedSecret);
    }
}
