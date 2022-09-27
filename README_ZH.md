# SealABC

[English](https://github.com/SealSC/SealABC_Consortium)

SealABC是一个高度灵活的模块化区块链系统开发框架。
它将区块链拆解为网络组件、共识组件、虚拟机组件、存储组件、应用组件，以及配套的工具集合。  
SealABC为不同组件制定了标准交互接口，并提供了这些组件的相应的实现，
让开发者可以直接使用SealABC框架提供的功能实现，快速搭建一条高性能区块链系统。  


## 框架特点

+ 标准功能集成：只需编写配置代码，就可以将不同标准组件组合为功能完备的高性能区块链系统。
+ 场景适配灵活：用户可以通过实现标准交互接口，可以提供自己实现的网络、共识、虚拟机等不同组件，灵活适配不同应用场景需求。
+ 应用扩展简洁：SealABC提供了一套应用层接口，让开发者无需关系区块链系统底层架构，专注于链上业务实现。



## 系统构建

SealABC 提供了一个默认的节点实现。 通过默认节点和相应的配置文件，开发者可以快速构建高性能的区块链系统。

1. 编译节点程序

```
./build.sh
    -p platform [linux|darwin]
    -a arch [amd64|arm]"
    -l path of config file

such as:
./build.sh -p linux -a amd64 -l ./config/example.json
```


2. 生成钱包

```
$ cd utils
$ go build walletgen.go

./walletgen <out_path> <option: privateKey>

such as:
./walletgen ./w1.json
```

3. 创建配置文件

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


4. 启动节点程序

```
./SealABC --config ./demo/node1/config-*.json -p <password>
```

## LICENSE

Apache 2.0
