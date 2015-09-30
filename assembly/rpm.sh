#首先打包引用

echo -------------------------------------------------------------------------
echo idcos-collect-verdb rpm 打包工具
echo -------------------------------------------------------------------------

#初始目录
CURRENT_DIR=`pwd`

#返回根目录
cd $CURRENT_DIR/..

#编译
. /etc/profile
gb build

#建立rpm包结构目录
rm -rfd temp
mkdir -p temp/bin

#copy打包之后的代码
cp bin/server temp/bin/verdb-server