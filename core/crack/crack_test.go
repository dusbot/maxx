package crack

import (
	"fmt"
	"testing"
)

func TestTelnet(t *testing.T) {
	c := new(TelnetCracker)
	c.Target = "10.1.2.138:23"
	c.User = "root"
	c.Pass = "root"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestHttp(t *testing.T) {
	c := new(HttpCracker)
	c.SetTarget("http://10.1.2.128:1080")
	// c.SetAuth("username", "password")
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestWmi(t *testing.T) {
	succ, err := WMIExec("10.1.2.137:135", "administrator", "adadministratormin", "", "", "", "", nil)
	fmt.Println(succ)
	fmt.Println(err)
}

func TestSnmp(t *testing.T) {
	c := new(SnmpCracker)
	c.SetTarget("10.1.2.138")
	c.SetAuth("", "public")
	succ, err := c.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println(succ)
}

func TestLdap(t *testing.T) {
	c := new(LdapCracker)
	c.Target = "10.1.2.216:389"
	c.User = "cn=admin,dc=nodomain"
	c.Pass = "administrator"
	succ, err := c.Crack()
	if err != nil {
		panic(err)
	}
	fmt.Println(succ)
}

func TestSmb(t *testing.T) {
	c := new(SmbCracker)
	c.Target = "10.1.2.137:445"
	c.User = "administrator"
	c.Pass = "administrator"
	succ, err := c.Crack()
	if err != nil {
		panic(err)
	}
	fmt.Println(succ)
}

func TestRsync(t *testing.T) {
	c := new(RsyncCracker)
	c.Target = "10.1.2.254:873"
	c.User = "root"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestSocks5(t *testing.T) {
	c := new(Socks5Cracker)
	c.Target = "10.1.2.128:1081"
	c.User = "myuser"
	c.Pass = "mypassword"
	succ, err := c.Ping()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestMssql(t *testing.T) {
	c := new(MssqlCracker)
	c.Target = "10.1.2.138:1433"
	c.User = "sa"
	c.Pass = "administrator"
	succ, err := c.Ping()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestOracle(t *testing.T) {
	c := new(OracleCracker)
	c.Target = "10.1.2.138:1521"
	c.User = "system"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestMqtt(t *testing.T) {
	c := new(MqttCracker)
	c.Target = "10.1.2.216:1883"
	c.User = "admin"
	c.Pass = "mqtt"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestMysql(t *testing.T) {
	c := new(MysqlCracker)
	c.Target = "10.1.2.138:3306"
	c.User = "root"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestRdp(t *testing.T) {
	c := new(RdpCracker)
	c.Target = "10.1.2.137:3389"
	c.User = "administrator"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestPostgres(t *testing.T) {
	c := new(PostgresCracker)
	c.Target = "10.1.2.138:5432"
	c.User = "postgres"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestAmqp(t *testing.T) {
	c := new(AmqpCracker)
	c.Target = "10.1.2.128:5672"
	c.User = "admin"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestVnc(t *testing.T) {
	c := new(VncCracker)
	c.Target = "10.1.2.137:5900"
	c.User = "admin"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestWinrm(t *testing.T) {
	c := new(WinrmCracker)
	c.Target = "10.1.2.137:5985"
	c.User = "administrator"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestRedis(t *testing.T) {
	c := new(RedisCracker)
	c.Target = "10.1.2.128:6379"
	c.User = ""
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestMemcached(t *testing.T) {
	c := new(MemcachedCracker)
	c.Target = "10.1.2.138:11211"
	c.User = ""
	c.Pass = ""
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestMonogodb(t *testing.T) {
	c := new(MongoCracker)
	c.Target = "10.1.2.128:27017"
	c.User = "root"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestTomcat(t *testing.T) {
	c := new(HttpCracker)
	c.Target = "http://10.1.2.128:11080/manager/html"
	c.User = "tomcat"
	c.Pass = "tomcat"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestJenkins(t *testing.T) {
	c := new(HttpCracker)
	c.Target = "10.1.1.45:8080"
	c.User = "administrator"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestGitlab(t *testing.T) {
	c := new(HttpCracker)
	c.Target = "10.1.1.57:9001"
	c.User = "administrator"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestNacos(t *testing.T) {
	c := new(HttpCracker)
	c.Target = "http://10.1.2.216:8848/nacos"
	c.User = "nacos"
	c.Pass = "nacos"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestNexus(t *testing.T) {
	c := new(HttpCracker)
	c.Target = "http://10.1.1.46:8081"
	c.User = "administrator"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}

func TestSvn(t *testing.T) {
	c := new(HttpCracker)
	c.Target = "http://192.168.10.23/cc/1.png"
	c.User = "administrator"
	c.Pass = "administrator"
	succ, err := c.Crack()
	fmt.Println(succ)
	fmt.Println(err)
}
