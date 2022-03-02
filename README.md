# Github backup

backup all github repo to local

### Install


```shell

npm i -g zx
wget https://raw.githubusercontent.com/tbxark/github-backup/master/backup.mjs
chmod +x backup.mjs

```

### Usage

```
./backup.mjs

#or

./backup.mjs --config=./you_config_path.json --target=./repos_save_path
```



### 碎碎念

最近看得到v2ex上有人github账户被封，俄罗斯也被米国盟友各种制裁。感觉把所有代码放在Github上不备份其实不太安全。
现有其他git平台不能实时同步github的仓库。得在每个github的repo添加action去同步，过于麻烦。
所以还是先备份到本地，然后再同步到另外的git平台。而且本地多一个备份也比较安全。
