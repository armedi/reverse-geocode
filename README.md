# reverse-geocode

Download geojson file here: 

You can generate your own data with higher resolution from this original [Shapefile](https://geoservices.big.go.id/portal/apps/webappviewer/index.html?id=49bda2cefd3f4b92aa726300bcdb40f7)

Set `GEOJSON_FILE` env with proper value.

Run the app

Hit `/area` endpoint

```sh
$ curl --request GET --url 'http://localhost:3000/area/-7.82306963629034,112.18998263914793'

{"FCODE":"BA03070040","KDBBPS":"3506","KDCBPS":"350601","KDCPUM":"35.06.08","KDEBPS":"3506011008","KDEPUM":"35.06.08.2006","KDPBPS":"35","KDPKAB":"35.06","KDPPUM":"35","LUAS":12.95209132,"LUASWH":0,"METADATA":"TASWIL1000020221227_DATA_BATAS_DESAKELURAHAN","NAMOBJ":"Gadungan","OBJECTID":10463,"REMARK":"","SRS_ID":"SRGI 2013","TIPADM":1,"UUPP":"Hasil Delineasi Batas Desa 2017","WADMKC":"Puncu","WADMKD":"Gadungan","WADMKK":"Kediri","WADMPR":"Jawa Timur","WIADKC":"","WIADKD":"","WIADKK":"","WIADPR":""}%
```