## Step1: install [Zombienet](https://github.com/paritytech/zombienet) 

## Step2: prepare relay chain  

```bash
# Clone
git clone https://github.com/paritytech/polkadot
cd polkadot

# Compile Polkadot with the real overseer feature
cargo build --release --bin polkadot
# put polkadot cli into your BIN PATH

``` 
## Step3: prepare parachain optional

```bash
# Clone
git clone https://github.com/paritytech/cumulus
cd cumulus

# Compile
cargo build --release --bin polkadot-parachain
# put polkadot-parachain cli into your BIN PATH

```

## Step4: spin up network 

```bash
# only relay chain 
zombienet spawn rococo.toml

# relay chain and parachain
zombienet spawn rococo-parachain.toml

```
