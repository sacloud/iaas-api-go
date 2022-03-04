# sacloud/iaas-api-go

- URL: https://github.com/sacloud/iaas-api-go/pull/2
- Parent: https://github.com/sacloud/sacloud-go/pull/1
- Author: @yamamoto-febc

## 概要

[sacloud-goの基本方針](https://github.com/sacloud/sacloud-go/pull/1)に従い、sacloud/libsacloud v2からIaaS部分を切り出す。

## やること/やらないこと

### やること

- libsacloudからのIaaS部分の切り出し
- iaas-api-go v1としてリリース
- libsacloudの`sacloud`パッケージ配下の整理
  - typesやostypeといったパッケージ構成の再考/整理
   
### やらないこと

- libsacloud v2の独自DSLを含むlibsacloudの実装の改善
  おおまか的にはlibsacloud v2をそのまま移植する。ただし、前述の`sacloud`パッケージ配下の整理などのリファクタレベルの修正は行う。
  従来[libsacloud v3として検討されてきた内容](https://github.com/sacloud/libsacloud/issues/791)はsacloud-goやiaas-api-go v2で実現する。

## 実装

TODO 切り出す範囲/修正する範囲を検討して記載

## 改訂履歴

- 2022/3/4: 初版作成