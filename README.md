# Titan

このプロジェクトでは、GeoFabrik社から提供されるOpenStreetMapのデータを使って、XYZベクトルタイルを生成するスクリプトを開発します。

## 概要

GeoFabrik社のダウンロードサイトからosm.pbf形式ファイルから、GeoJSON形式ファイルへへ変換します。

## 利用方法

+ 現時点では、四国（shikoku-latest.osm.pbf）の学校データ（タグ名amenity=school）のみ動作検証済みです。  
+ 変換対象の図形種別は、Point、Polygon、MultiPolygonの３種類です。  
+ osm.pbfのNode、Way、Relationとも対応しています。
+ 事前にosmiumにて、NodeとWay間の関係付けが必要です。コマンドイメージは以下の通りです。

`% osmium add-locations-to-ways -n -o shikoku-low.osm.pbf shikoku-latest.osm.pbf`
