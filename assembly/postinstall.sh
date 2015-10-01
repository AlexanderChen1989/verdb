chmod 755 /usr/yunji/idcos-collect-verdb/bin/verdb-server
chown root:root /usr/yunji/idcos-collect-verdb/bin/verdb-server

if [ `readlink -f /usr/local/bin/verdb-server` != "/usr/yunji/idcos-collect-verdb/bin/verdb-server" ]; then
	rm -rf /usr/local/bin/verdb-server
	ln -s /usr/yunji/idcos-collect-verdb/bin/verdb-server /usr/local/bin/verdb-server
fi

service mongod status || service mongod start
/usr/local/bin/verdb-server &>/var/log/verdb-server.log &