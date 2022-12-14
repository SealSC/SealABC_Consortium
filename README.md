# SealABC

[中文](https://github.com/SealSC/SealABC_Consortium/blob/master/README_ZH.md)

SealABC is a highly flexible modular blockchain system development framework. 
It disassembles the blockchain into network components, consensus components, virtual machine components, storage components, application components, and a set of utility tools.
SealABC formulates standard interaction interfaces for different components and provides corresponding implementations of these components, allowing developers to directly use the components provided by the SealABC framework to quickly build a high-performance blockchain system.

## Features

+ Standard components integration: Different standard components can be combined into a fully functional high-performance blockchain system by simply writing configuration code. 
+ Flexible scenario adaptation: Users can provide their own network, consensus, virtual machine and other components to flexibly adapt to the needs of different application scenarios through implementing standard interaction interfaces
+ Simple application expansion: SealABC provides a set of application layer interfaces, which allows developers to focus on the implementation of on-chain business without concerning with the underlying architecture of the blockchain system.



## System build

SealABC provides a default node implementation. With default nodes and corresponding configuration files, developers can quickly build high-performance blockchain systems.

1. build node

```
./build.sh
    -p platform [linux|darwin]
    -a arch [amd64|arm]"
    -l path of config file

such as:
./build.sh -p linux -a amd64 -l ./config/example.json
```


2. generate wallet

```
$ cd utils
$ go build walletgen.go

./walletgen <out_path> <option: privateKey>

such as:
./walletgen ./w1.json
```

3. make config files

```
{
  "consensus_conf": {
    "consensus_disabled": false,
    "consensus_service_protocol": "tcp",
    "consensus_service_address": "127.0.0.1:30001",
    "consensus_member": [
      "127.0.0.1:30101",
      "127.0.0.1:30201",
      "127.0.0.1:30301",
      "127.0.0.1:30401"
    ],
    "members": [
      "039bc024533c28827c13c2df0d546283546fd2da033d7ed9b1f22d4c4906a58cbc",
      "0201d331ddcd43d881d9f1b6865d76e27c95dc96ad61118513eaaa0cb94b86530f",
      "029da5d3be11f43ca8d23bc13ab199ad2bda4d8aa8050b8dbe3f74afe2bf2b52a1",
      "02452ed551c3d9e57f0ea742997d9bf3efb5c0fbe535050cce55120967233c8c57",
      "03f75b01f91f4a7438d7c7c693b46f13f79f2b9010e1a1b77e442c2694dd85277c"
    ],
    "online_check_interval": 1000,
    "consensus_timeout": 10000,
    "consensus_Interval": 3000
  },
  "block_chain_conf": {
    "blockchain_service_address": "127.0.0.1:30002",
    "blockchain_service_seeds": [
      "127.0.0.1:30102",
      "127.0.0.1:30202",
      "127.0.0.1:30302",
      "127.0.0.1:30402"
    ],
    "blockchain_api_config": {
      "address": ":30003",
      "base_path": "/api/v1"
    },
    "chain_db": "./demo/node1/db/chain",
    "blockchain_service_protocol": "tcp"
  },
  "debug_conf": {
    "pprof_port": "localhost:6060"
  },
  "log_conf": {
    "log_file": "./demo/node1/current.log",
    "log_level": 5
  },
  "wallet_conf": {
    "wallet_path": "./demo/node1/wallet1.json"
  },
  "utxo_app_conf": {
    "utxo_app_enable": true,
    "utxo_ledger_db": "./demo/node1/db/ledger"
  },
  "memo_app_conf": {
    "memo_app_enable": true,
    "memo_db": "./demo/node1/db/memo"
  },
  "smart_assets_app_conf": {
    "smart_assets_enable": true,
    "smart_assets_db": "./demo/node1/db/smart_assets",
    "tx_pool_limit": 1000,
    "client_tx_limit": 100
  },
  "mysql_conf": {
    "enable_sql_storage": false
  },
  "engine_conf": {
    "engine_api_config": {
      "address": ":30004",
      "base_path": "/api/v1"
    }
  },
  "crypto_conf": {
    "hash_type": "keccak_256",
    "cipher_type": "AES",
    "signer_type": "secp256k1"
  }
}

```


4. startup nodes

```
./SealABC --config ./demo/node1/config-*.json -p <password>
```

## LICENSE

Apache 2.0
