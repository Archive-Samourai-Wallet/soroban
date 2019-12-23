package com.samouraiwallet.soroban.client;

import java.io.IOException;
import java.security.NoSuchAlgorithmException;
import java.time.LocalDateTime;
import java.util.concurrent.TimeoutException;

import com.samouraiwallet.soroban.rpc.RpcClient;
import com.samouraiwallet.soroban.User;
import com.samouraiwallet.soroban.Box;

import org.apache.commons.codec.DecoderException;

public final class SorobanClient {

    public static void main(String[] args) throws IOException {

        // parse cli
        String torProxy = "127.0.0.1";
        int torPort = 9050;
        String torUrl = "http://sorzvujomsfbibm7yo3k52f3t2bl6roliijnm7qql43bcoe2kxwhbcyd.onion";
        String role = "contributor";
        String directoryName = "samourai.soroban.private";
        int numInter = 3;

        if (args.length > 0){
            torProxy = args[0];
        }
        if (args.length > 1){
            torPort = Integer.parseInt(args[1]);
        }
        if (args.length > 2){
            torUrl = args[2];
        }
        if (args.length > 3){
            role = args[3];
        }
        if (args.length > 4){
            directoryName = args[4];
        }
        if (args.length > 5){
            numInter = Integer.parseInt(args[5]);
        }

        RpcClient rpc = new RpcClient(torProxy, torPort, torUrl);
        try {
            Boolean hashEncode = true;
            if (hashEncode) {
                directoryName = RpcClient.encodeDirectory(directoryName);   
            }

            while (true) {
                switch(role) {
                    case "initiator": SorobanClient.initiator(rpc, directoryName, hashEncode, numInter); break;
                    case "contributor": SorobanClient.contributor(rpc, directoryName, hashEncode, numInter); break;
                }
            }

        } catch (NoSuchAlgorithmException e) {
            e.printStackTrace();
        } finally {
            rpc.close();
        }
    }

    static void initiator(RpcClient rpc, String directoryName, Boolean hashEncode, int numIter) {
        User user = new User();
        try{
            System.out.println("Registering public_key");
            rpc.directoryAdd(directoryName, user.publicKey(), "long");

            String privateDirectory = String.format("%s.%s", directoryName, user.publicKey());
            if (hashEncode) {
                privateDirectory = RpcClient.encodeDirectory(privateDirectory);   
            }

            String candidatePublicKey = rpc.waitAndRemove(privateDirectory, 100);
            if (candidatePublicKey.isEmpty()){
                System.out.println("Invalid candidate_publicKey");
                
            }
            System.out.println("candidate publicKey found");
            System.out.println(candidatePublicKey);

            Box box = user.box(candidatePublicKey);
            String nextDirectory = RpcClient.encodeDirectory(user.sharedSecret(box));
            if (nextDirectory.isEmpty()){
                System.out.println("Invalid reponse next_directory");
                return;
            }
            System.out.printf("nextDirectory: %s%n", nextDirectory);

            String payload;
            System.out.println("Starting echange loop...");
            int counter = 1;
            while (numIter>0) {
                // request
                System.out.println("Sending : Ping");
                payload = box.encrypt(String.format("Ping %d %s", counter, LocalDateTime.now()));
                if (payload.isEmpty()) {
                    System.out.println("Invalid Ping message");
                    return;
                }

                rpc.directoryAdd(nextDirectory, payload, "short");
                nextDirectory = RpcClient.encodeDirectory(payload);
                if (nextDirectory.isEmpty()){
                    System.out.println("Invalid reponse next_directory");
                    return;
                }

                // response
                payload = rpc.waitAndRemove(nextDirectory, 10);
                nextDirectory = RpcClient.encodeDirectory(payload);
                if (nextDirectory.isEmpty()) {
                    System.out.println("Invalid reponse next_directory");
                    return;
                }

                String message = box.decrypt(payload);
                if (message.isEmpty()){
                    System.out.println("Invalid reponse message");
                    return;
                }
                System.out.println(String.format("Recieved: %s", message));
                counter += 1;
                numIter -= 1;
            }
                
        } catch(NoSuchAlgorithmException e) {
            e.printStackTrace();
        } catch(DecoderException e) {
            e.printStackTrace();
        } catch(TimeoutException e) {
            e.printStackTrace();
        } catch(InterruptedException e) {
            e.printStackTrace();
        } catch(IOException e) {
            e.printStackTrace();
        }
    }

    static void contributor(RpcClient rpc, String directoryName, Boolean hashEncode, int numIter) {
        User user = new User();
        try{
            System.out.println("Waiting for initiator publicKey");
            String initiatorPublicKey = rpc.waitAndRemove(directoryName, 10);
            if (initiatorPublicKey.isEmpty()){
                System.out.println("Invalid initiator publicKey");
            }
            System.out.println("Initiator publicKey found");
            String privateDirectory = String.format("%s.%s", directoryName, initiatorPublicKey);
            if (hashEncode) {
                privateDirectory = RpcClient.encodeDirectory(privateDirectory);   
            }

            System.out.println("Sending public_key");
            rpc.directoryAdd(privateDirectory, user.publicKey(), "default");

            Box box = user.box(initiatorPublicKey);
            String nextDirectory = RpcClient.encodeDirectory(user.sharedSecret(box));
            if (nextDirectory.isEmpty()){
                System.out.println("Invalid next_directory (start)");
                return;
            }

            String payload;
            System.out.println("Starting echange loop...");
            int counter = 1;
            while (numIter>0) {
                // query
                payload = rpc.waitAndRemove(nextDirectory, 10);
                if (payload.isEmpty()) {
                    System.out.println("Invalid payload (query)");
                    return;
                }
                nextDirectory = RpcClient.encodeDirectory(payload);
                if (nextDirectory.isEmpty()) {
                    System.out.println("Invalid next_directory (query)");
                    return;
                }

                String message = box.decrypt(payload);
                if (message.isEmpty()){
                    System.out.println("Invalid reponse message");
                    return;
                }

                // request
                System.out.println("Sending : Ping");
                payload = box.encrypt(String.format("Ping %d %s", counter, LocalDateTime.now()));
                if (payload.isEmpty()) {
                    System.out.println("Invalid query");
                    return;
                }
                System.out.printf("Recieved: %s%n", message);

                // response
                System.out.printf("Replying: %s%n", "Pong");
                payload = box.encrypt(String.format("Pong %d %s", counter, LocalDateTime.now()));
                if (payload.isEmpty()) {
                    System.out.println("Invalid payload (reply)");
                    return;
                }
                rpc.directoryAdd(nextDirectory, payload, "short");

                nextDirectory = RpcClient.encodeDirectory(payload);
                if (nextDirectory.isEmpty()) {
                    System.out.println("Invalid next_directory (response)");
                    return;
                }

                counter += 1;
                numIter -= 1;
            }
                
        } catch(NoSuchAlgorithmException e) {
            e.printStackTrace();
        } catch(DecoderException e) {
            e.printStackTrace();
        } catch(TimeoutException e) {
            e.printStackTrace();
        } catch(InterruptedException e) {
            e.printStackTrace();
        } catch(IOException e) {
            e.printStackTrace();
        }
    }
}
