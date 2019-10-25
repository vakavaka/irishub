# TEST

## Install `iris` `gaia` `relayer`

Install `iris`

```bash
git clone https://github.com/irisnet/irishub.git
cd irishub
git checkout cosmos-sdk
go mod tidy && make install
```

Install `gaia`

```bash
git clone https://github.com/irisnet/gaia.git
cd gaia
git checkout ibc-gaia-mockbank
go mod tidy && make install
```

Install `relayer`

```bash
git clone https://github.com/chengwenxi/cosmos-relayer.git
cd cosmos-relayer
go mod tidy && make install
```

## Initiate

```bash
chmod 777 init.sh
./init.sh
```

## Start `iris` and `gaia`

```bash
nohup iris --home ibc-iris/n0/iris start >ibc-iris.log &
nohup gaiad --home ibc-gaia/n0/gaiad start >ibc-gaia.log &
```

## Handshake

Create `client`, `connection` handshake, and `channel` handshake

```bash
chmod 777 handshake.sh
handshake.sh
```

## Transfer test

Transfer from `iris`

```bash
chmod 777 transfer-iris-test.sh
transfer-iris-test.sh
```

Transfer from `gaia`

```bash
chmod 777 transfer-gaia-test.sh
transfer-gaia-test.sh
```

## Transfer with relayer

Start `relayer`

```bash
relayer start \
    "iris" "tcp://localhost:26657" "n0" "12345678" "ibc-iris/n0/iriscli/" "client-to-gaia" \
    "cosmos" "tcp://localhost:26557" "n0" "12345678" "ibc-gaia/n0/gaiacli/" "client-to-iris"
```

Transfer from `iris`

```bash
chmod 777 transfer-iris.sh
transfer-iris.sh
```

Transfer from `gaia`

```bash
chmod 777 transfer-gaia.sh
transfer-gaia.sh
```

Query result

```bash
chmod 777 query-accounts.sh
query-accounts.sh
```
