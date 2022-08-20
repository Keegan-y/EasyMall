BUILD_ID=DONTKILLME
srv_name="goods_web_main"
chmod +x ./$srv_name
#重启,如果已经存在则先关闭再重启
if pgrep -x $srv_name > /dev/null
then
  echo "${srv_name} is running"
  echo "shutting down ${srv_name}"
  if ps -a | grep $srv_name | awk '{print $1}' | xargs kill $1
    then
      echo "starting ${srv_name}"
      ./$srv_name > /dev/null 2>&1 &
      echo "start ${srv_name} success"
  fi
else
  echo "starting ${srv_name}"
  ./$srv_name > /dev/null 2>&1 &
  echo "start ${srv_name} success"
fi