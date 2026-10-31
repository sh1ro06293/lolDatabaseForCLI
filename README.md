# loldatabase

League of Legends のチャンピオン情報を日本語で表示する Windows 向けの CLI です。開発中のため、まだ配布していません。

## 免責

- loldatabase は個人が作った非公式のツールです。使って起きた結果（アカウントの停止、データの消失など）について、作者は責任を負いません。詳しくは [LICENSE](LICENSE) を参照してください。
- このツールはゲームのクライアントに触れず、ゲームのファイルも書き換えません。インターネットで公開されているデータを読んで表示するだけです。
- LICENSE が対象にするのは、このリポジトリのコードと文書です。取得して表示するチャンピオンの名前、スキルの説明や数値などのゲームのデータは対象外で、Riot Games のものです。

loldatabase isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties. Riot Games, and all associated properties are trademarks or registered trademarks of Riot Games, Inc.

## 開発

### 確認

```sh
task check
```

### チャンピオンデータの構造の監査

Riot Data Dragon の日本語チャンピオン一覧を基準に、同じパッチの CommunityDragon に日本語のパッシブ・Q/W/E/R とゲーム内スキルレコードがあるか確認します。特殊モード専用のレコードは Riot の一覧にないため対象外です。

```sh
go run ./cmd/data-audit
```

Data Dragon の最新版を取得して監査します。過去の版を監査するときは `-version <Data Dragon の版>` を付けます。結果は JSON で出力し、構造上の欠落があれば終了コード 1、取得開始に失敗した場合は終了コード 2 です。通信を行い、取得内容は保存しません。

この検証はファイルの存在と基本構造だけを調べます。ダメージ・回復量などの計算式、変身後のスキル、パッチの内容が正しいことは**まだ保証しません**。そのため、このコマンドの成功だけをリリース許可条件にはできません。
