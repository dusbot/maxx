package scan

var DefaultPorts = []int{
	7,     // Echo
	19,    // Chargen
	20,    // FTP Data
	21,    // FTP Control
	22,    // SSH
	23,    // Telnet
	25,    // SMTP
	37,    // Time
	42,    // WINS
	43,    // WHOIS
	49,    // TACACS
	53,    // DNS
	67,    // DHCP Server
	68,    // DHCP Client
	69,    // TFTP
	70,    // Gopher
	79,    // Finger
	80,    // HTTP
	88,    // Kerberos
	102,   // S7
	110,   // POP3
	111,   // RPCbind
	113,   // Ident
	119,   // NNTP
	123,   // NTP
	135,   // MS RPC
	137,   // NetBIOS-NS
	138,   // NetBIOS-DGM
	139,   // NetBIOS-SSN
	143,   // IMAP
	161,   // SNMP
	162,   // SNMP Trap
	177,   // XDMCP
	179,   // BGP
	194,   // IRC
	199,   // SMUX
	201,   // AppleTalk
	264,   // BGMP
	318,   // TSP
	381,   // HP OpenView
	383,   // HP OpenView
	389,   // LDAP
	411,   // Direct Connect
	412,   // Direct Connect
	427,   // SLP
	443,   // HTTPS
	445,   // SMB
	464,   // Kerberos
	465,   // SMTPS
	497,   // Dantz Retrospect
	500,   // IPsec/ISAKMP
	502,   // Modbus
	512,   // Rexec
	513,   // Rlogin
	514,   // Syslog
	515,   // LPD
	520,   // RIP
	521,   // RIPng
	540,   // UUCP
	543,   // Kerberos
	544,   // kshell
	546,   // DHCPv6 Client
	547,   // DHCPv6 Server
	548,   // AFP
	554,   // RTSP
	563,   // NNTPS
	587,   // SMTP Submission
	591,   // FileMaker
	593,   // MS RPC over HTTP
	631,   // IPP
	636,   // LDAPS
	639,   // MSDP
	646,   // LDP
	691,   // MS Exchange
	860,   // iSCSI
	873,   // rsync
	902,   // VMware
	989,   // FTPS Data
	990,   // FTPS Control
	993,   // IMAPS
	995,   // POP3S
	1025,  // NFS or others
	1026,  // Windows Events
	1029,  // MS RPC
	1080,  // SOCKS
	1099,  // Java RMI
	1194,  // OpenVPN
	1214,  // Kazaa
	1241,  // Nessus
	1311,  // Dell OpenManage
	1337,  // WASTE
	1433,  // MSSQL
	1434,  // MSSQL Monitor
	1512,  // WINS
	1521,  // Oracle
	1589,  // Cisco VQP
	1604,  // Citrix
	1645,  // RADIUS
	1646,  // RADIUS Accounting
	1701,  // L2TP
	1720,  // H.323
	1723,  // PPTP
	1755,  // MMS
	1812,  // RADIUS
	1813,  // RADIUS Accounting
	1863,  // MSN
	1900,  // UPnP
	1935,  // RTMP
	1984,  // Big Brother
	2000,  // SCCP
	2002,  // Cisco ACS
	2049,  // NFS
	2082,  // cPanel
	2083,  // cPanel SSL
	2100,  // Oracle XDB
	2222,  // DirectAdmin
	2301,  // Compaq Insight
	2323,  // 3DNF
	2381,  // etcd
	2404,  // IEC 60870-5-104
	2427,  // Media Gateway
	2483,  // Oracle DB SSL
	2484,  // Oracle DB
	2546,  // VytalVault
	2967,  // Symantec AV
	3000,  // Node.js
	3050,  // Interbase
	3074,  // Xbox Live
	3128,  // Squid
	3260,  // iSCSI Target
	3306,  // MySQL
	3389,  // RDP
	3396,  // Novell NDPS
	3689,  // DAAP
	3690,  // SVN
	3724,  // World of Warcraft
	3784,  // Ventrilo
	3785,  // Ventrilo
	4000,  // Custom
	4063,  // EPICS
	4064,  // EPICS
	4100,  // WatchGuard
	4333,  // mSQL
	4444,  // Metasploit
	4500,  // IPsec NAT-T
	4567,  // Sinatra
	4662,  // eMule
	4672,  // eMule
	4899,  // Radmin
	5000,  // Custom
	5050,  // Yahoo! Messenger
	5060,  // SIP
	5190,  // AIM
	5222,  // XMPP
	5223,  // XMPP SSL
	5269,  // XMPP Server
	5353,  // mDNS
	5432,  // PostgreSQL
	5500,  // VNC
	5555,  // Android ADB
	5601,  // Kibana
	5631,  // pcAnywhere
	5666,  // Nagios
	5800,  // VNC HTTP
	5900,  // VNC
	5938,  // TeamViewer
	5984,  // CouchDB
	5985,  // WinRM HTTP
	5986,  // WinRM HTTPS
	6000,  // X11
	6001,  // X11
	6379,  // Redis
	6666,  // IRC
	6667,  // IRC
	6668,  // IRC
	6669,  // IRC
	6679,  // IRC SSL
	6697,  // IRC SSL
	6881,  // BitTorrent
	6969,  // BitTorrent
	7000,  // Azureus
	7001,  // WebLogic
	7002,  // WebLogic
	7070,  // RealServer
	7080,  // Play!
	7100,  // Font Service
	7547,  // CPE WAN
	7777,  // Oracle
	8000,  // Custom
	8005,  // Tomcat
	8008,  // HTTP Alt
	8009,  // Tomcat AJP
	8010,  // XMPP File Transfer
	8080,  // HTTP Alt
	8081,  // HTTP Alt
	8088,  // HTTP Alt
	8090,  // HTTP Alt
	8091,  // HTTP Alt
	8099,  // HTTP Alt
	8100,  // HTTP Alt
	8181,  // HTTP Alt
	8200,  // GoCD
	8222,  // VMware
	8243,  // HTTPS Alt
	8280,  // HTTP Alt
	8333,  // Bitcoin
	8400,  // Commvault
	8443,  // HTTPS Alt
	8500,  // HashiCorp Consul
	8530,  // Windows Server Update
	8531,  // Windows Server Update
	8888,  // HTTP Alt
	8983,  // Apache Solr
	9000,  // Custom
	9001,  // Custom
	9042,  // Apache Cassandra
	9060,  // WebSphere
	9080,  // WebSphere
	9090,  // HTTP Alt
	9091,  // HTTP Alt
	9100,  // JetDirect
	9200,  // Elasticsearch
	9300,  // Elasticsearch
	9418,  // Git
	9443,  // HTTPS Alt
	9500,  // ISCSI
	9535,  // mRemoteNG
	9600,  // Logstash
	9675,  // Spice
	9695,  // CCNx
	9876,  // Custom
	9999,  // Custom
	10000, // Webmin
	10001, // Custom
	10050, // Zabbix
	10051, // Zabbix
	10123, // NetIQ
	10250, // Kubernetes
	11211, // Memcached
	12345, // NetBus
	13720, // NetBackup
	13721, // NetBackup
	13782, // NetBackup
	13783, // NetBackup
	15118, // Dell OpenManage
	16992, // Intel AMT
	16993, // Intel AMT
	18080, // HTTP Alt
	19132, // Minecraft
	19283, // Kaseya
	20000, // DNP3
	20547, // ProConOS
	21025, // Starbound
	22000, // Custom
	22136, // FLIR
	22222, // Custom
	23023, // Telnet Alt
	23424, // Custom
	25565, // Minecraft
	26000, // Custom
	27015, // Steam
	27017, // MongoDB
	27018, // MongoDB
	27374, // Sub7
	28960, // Call of Duty
	31337, // Back Orifice
	32400, // Plex
	32764, // Linux Backdoor
	33434, // traceroute
	37777, // Custom
	40000, // Custom
	47808, // BACnet
	49152, // Windows RPC
	50000, // Custom
	50030, // Hadoop
	50060, // Hadoop
	50070, // Hadoop
	54321, // BO2K
	55055, // Custom
	55553, // Metasploit
	57722, // Custom
	60010, // Hadoop
	64738, // Mumble
}