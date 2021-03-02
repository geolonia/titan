# Titan

このプロジェクトでは、GeoFabrik社（https://www.geofabrik.de）から提供されるOpenStreetMapのデータを使って、XYZベクトルタイルを生成するスクリプトを開発します。

## 概要

GeoFabrik社が定期的に更新するダウンロードサイト（http://download.geofabrik.de）からosmのXML形式ファイルを入手し、mbtilesへ変換します。

## 利用方法

現時点では、住所データの抽出のみ実現しています。
実行方法は、titan.yamlに抽出タグ名と出力ファイル名を指定し、getnode.pyを実行します。