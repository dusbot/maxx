package crack

import (
	"errors"
)

var ERR_CONNECTION = errors.New("connection failed")

type Crack interface {
	Ping() (bool, error)
	Crack() (bool, error)

	SetTarget(target string)
	SetIpPort(ip, port string)
	SetService(service string)
	SetTimeout(timeout int)
	SetAuth(user, pass string)
}

type CrackBase struct {
	Service, Target, Ip, Port, User, Pass string
	Timeout                               int
}

func (c *CrackBase) SetTarget(target string) {
	c.Target = target
}

func (c *CrackBase) SetIpPort(ip, port string) {
	c.Ip = ip
	c.Port = port
}

func (c *CrackBase) SetService(service string) {
	c.Service = service
}

func (c *CrackBase) SetAuth(user, pass string) {
	c.User = user
	c.Pass = pass
}

func (c *CrackBase) SetTimeout(timeout int) {
	c.Timeout = timeout
}

const (
	CRACK_FTP           = "FTP"
	CRACK_SSH           = "SSH"
	CRACK_TELNET        = "TELNET"
	CRACK_HTTPBASIC     = "HTTP"
	CRACK_WMI           = "WMI"
	CRACK_SNMP          = "SNMP"
	CRACK_LDAP          = "LDAP"
	CRACK_SMB           = "SMB"
	CRACK_RTSP          = "RTSP"
	CRACK_RSYNC         = "RSYNC"
	CRACK_SOCKS5        = "SOCKS5"
	CRACK_MSSQL         = "MSSQL"
	CRACK_ORACLE        = "ORACLE"
	CRACK_MQTT          = "MQTT"
	CRACK_MYSQL         = "MYSQL"
	CRACK_RDP           = "RDP"
	CRACK_POSTGRESQL    = "POSTGRESQL"
	CRACK_AMQP          = "AMQP"
	CRACK_VNC           = "VNC"
	CRACK_WINRM         = "WINRM"
	CRACK_REDIS         = "REDIS"
	CRACK_MEMCACHED     = "MEMCACHED"
	CRACK_MONGODB       = "MONGODB"
	CRACK_TOMCAT        = "TOMCAT"
	CRACK_JENKINS       = "JENKINS"
	CRACK_GITLAB        = "GITLAB"
	CRACK_NACOS         = "NACOS"
	CRACK_NEXUS         = "NEXUS"
	CRACK_SVN           = "SVN"
	CRACK_ELASTICSEARCH = "ELASTICSEARCH"
	CRACK_WEBLOGIC      = "WEBLOGIC"
	CRACK_EXPRESS       = "EXPRESS"
	CRACK_HABASE_REST   = "HBASE_REST_API"
	CRACK_FLASK         = "FLASK"
	CRACK_GIN           = "GIN"
	CRACK_PROMETHEUS    = "PROMETHEUS"
	CRACK_APACHE        = "APACHE"
	CRACK_GRAFANA       = "GRAFANA"
	CRACK_MINIO         = "MINIO"
	CRACK_ENVOY         = "ENVOY"
)

type CrackTemplate func() Crack

var (
	CrackServiceMap = map[string]CrackTemplate{
		CRACK_FTP: func() Crack {
			return &FtpCracker{}
		},
		CRACK_SSH: func() Crack {
			return &SshCracker{}
		},
		CRACK_TELNET: func() Crack {
			return &TelnetCracker{}
		},
		CRACK_HTTPBASIC: func() Crack {
			return &HttpCracker{}
		},
		CRACK_WMI: func() Crack {
			return &WmiCracker{}
		},
		CRACK_SNMP: func() Crack {
			return &SnmpCracker{}
		},
		CRACK_LDAP: func() Crack {
			return &LdapCracker{}
		},
		CRACK_SMB: func() Crack {
			return &SmbCracker{}
		},
		CRACK_RSYNC: func() Crack {
			return &RsyncCracker{}
		},
		CRACK_SOCKS5: func() Crack {
			return &Socks5Cracker{}
		},
		CRACK_MSSQL: func() Crack {
			return &MssqlCracker{}
		},
		CRACK_ORACLE: func() Crack {
			return &OracleCracker{}
		},
		CRACK_MQTT: func() Crack {
			return &MqttCracker{}
		},
		CRACK_MYSQL: func() Crack {
			return &MysqlCracker{}
		},
		CRACK_RDP: func() Crack {
			return &RdpCracker{}
		},
		CRACK_POSTGRESQL: func() Crack {
			return &PostgresCracker{}
		},
		CRACK_AMQP: func() Crack {
			return &AmqpCracker{}
		},
		CRACK_VNC: func() Crack {
			return &VncCracker{}
		},
		CRACK_WINRM: func() Crack {
			return &WinrmCracker{}
		},
		CRACK_REDIS: func() Crack {
			return &RedisCracker{}
		},
		CRACK_MEMCACHED: func() Crack {
			return &MemcachedCracker{}
		},
		CRACK_MONGODB: func() Crack {
			return &MongoCracker{}
		},
		CRACK_TOMCAT: func() Crack {
			return &HttpCracker{}
		},
		CRACK_JENKINS: func() Crack {
			return &HttpCracker{}
		},
		CRACK_GITLAB: func() Crack {
			return &HttpCracker{}
		},
		CRACK_NACOS: func() Crack {
			return &HttpCracker{}
		},
		CRACK_NEXUS: func() Crack {
			return &HttpCracker{}
		},
		CRACK_SVN: func() Crack {
			return &HttpCracker{}
		},
		CRACK_ELASTICSEARCH: func() Crack {
			return &ElasticsearchCracker{}
		},
		CRACK_WEBLOGIC: func() Crack {
			return &HttpCracker{}
		},
		CRACK_EXPRESS: func() Crack {
			return &HttpCracker{}
		},
		CRACK_HABASE_REST: func() Crack {
			return &HttpCracker{}
		},
		CRACK_FLASK: func() Crack {
			return &HttpCracker{}
		},
		CRACK_GIN: func() Crack {
			return &HttpCracker{}
		},
		CRACK_PROMETHEUS: func() Crack {
			return &HttpCracker{}
		},
		CRACK_APACHE: func() Crack {
			return &HttpCracker{}
		},
		CRACK_GRAFANA: func() Crack {
			return &HttpCracker{}
		},
		CRACK_MINIO: func() Crack {
			return &HttpCracker{}
		},
		CRACK_ENVOY: func() Crack {
			return &HttpCracker{}
		},
	}
	DefaultPortService = map[int]string{
		21:    CRACK_FTP,
		22:    CRACK_SSH,
		23:    CRACK_TELNET,
		80:    CRACK_HTTPBASIC,
		443:   CRACK_HTTPBASIC,
		135:   CRACK_WMI,
		161:   CRACK_SNMP,
		389:   CRACK_LDAP,
		445:   CRACK_SMB,
		554:   CRACK_RTSP,
		873:   CRACK_RSYNC,
		1080:  CRACK_SOCKS5,
		1433:  CRACK_MSSQL,
		1521:  CRACK_ORACLE,
		1883:  CRACK_MQTT,
		3306:  CRACK_MYSQL,
		3389:  CRACK_RDP,
		5432:  CRACK_POSTGRESQL,
		5672:  CRACK_AMQP,
		5900:  CRACK_VNC,
		5985:  CRACK_WINRM,
		5986:  CRACK_WINRM,
		6379:  CRACK_REDIS,
		11211: CRACK_MEMCACHED,
		27017: CRACK_MONGODB,
		8080:  CRACK_TOMCAT,
		8081:  CRACK_JENKINS,
		8082:  CRACK_NACOS,
		8083:  CRACK_NEXUS,
		3690:  CRACK_SVN,
		8084:  CRACK_GITLAB,
		8443:  CRACK_GITLAB,
	}
)
