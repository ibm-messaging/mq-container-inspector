#### Prerequisites
To perform signature verification, your machine must have the following command line tools installed:

[Gnu Privacy Guard v2](https://gnupg.org/)

### Steps to verify MQ container inspector tarball

1. Download the MQ public key

Run one of the following commands to download the MQ public key in a file named mq-public.gpg:

Retrieve the latest public keys (example with wget):

`wget https://raw.githubusercontent.com/ibm-messaging/mq-container-inspector/refs/heads/main/certificates/mq-public.gpg`

Retrieve the latest public keys (example with curl):

`curl https://raw.githubusercontent.com/ibm-messaging/mq-container-inspector/refs/heads/main/certificates/mq-public.gpg -o mq-public.gpg`

2. Import the MQ container public key to your machine

Note: This step only needs to be done once on each machine used for signature verification.

```
gpg --import mq-public.gpg
```

You should see a message like the following. 

```
gpg: key 951404A6E7853923: "International Business Machines Corporation <psirt@us.ibm.com>" imported
```

Please verify your key id matches "951404A6E7853923". 

3. Generate a gpg secret key

To trust the MQ public key, you need to sign it with your secret key. 
If you do not already have a gpg private key, Run the command below to generate one:

```
gpg --generate-key
```

You will be asked to provide a `Real name` and `Email address` to identify the owner of the key.

4. Sign the MQ container public key

By signing the MQ container public key with your secret key, you are marking it as a trusted key.

Run the following command using the MQ public key if from step 2:

```
gpg --lsign-key 951404A6E7853923
```

5. Verify the tarball

Run the following command to verify the tarball

```
gpg --verify <signature> <tarball> 
```

Where:
  - `<signature>` is the `.asc` signature file (e.g. 1.0.0_mq-container-inspector_linux-amd64.tar.gz.asc)
  - `<tarball>` is the signed folder (e.g. 1.0.0_mq-container-inspector_linux-amd64.tar.gz)

4. Confirm the signature message is correct

If the signature is valid you should see a successfull validation message like the one below:

```
gpg: Signature made Fri  3 Oct 12:34:04 2025 BST
gpg:                using RSA key 81F0780ACF05AC8A7017A22B951404A6E7853923
gpg: checking the trustdb
gpg: marginals needed: 3  completes needed: 1  trust model: pgp
gpg: depth: 0  valid:   1  signed:   1  trust: 0-, 0q, 0n, 0m, 0f, 1u
gpg: depth: 1  valid:   1  signed:   0  trust: 0-, 0q, 0n, 0m, 1f, 0u
gpg: next trustdb check due at 2028-10-03
gpg: Good signature from "International Business Machines Corporation <psirt@us.ibm.com>" [full]
```