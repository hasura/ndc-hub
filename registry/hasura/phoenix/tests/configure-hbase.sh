#!/bin/bash

echo 'Configuring HBase for standalone mode...'

cat > /opt/hbase/conf/hbase-site.xml << 'EOF'
<?xml version="1.0"?>
<configuration>
  <property>
    <name>hbase.rootdir</name>
    <value>file:///tmp/hbase</value>
  </property>
  <property>
    <name>hbase.zookeeper.property.dataDir</name>
    <value>/tmp/zookeeper</value>
  </property>
  <property>
    <name>hbase.unsafe.stream.capability.enforce</name>
    <value>false</value>
  </property>
  <property>
    <name>hbase.cluster.distributed</name>
    <value>false</value>
  </property>
</configuration>
EOF

echo 'HBase configuration complete!'
