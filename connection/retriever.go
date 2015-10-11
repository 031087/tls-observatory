package connection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os/exec"
)

var cipherscan string

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func panicIf(err error) {
	if err != nil {
		log.Println(fmt.Sprintf("%s", err))
	}
}

type NoTLSConnErr string

func (f NoTLSConnErr) Error() string {
	return fmt.Sprintf("No TLS Certs Received from %s", string(f))
}

func Connect(domain string) ([]byte, error) {

	ip := getRandomIP(domain)

	if ip == "" {
		e := fmt.Errorf("Could not resolve ip for: ", domain)
		log.Println(e)
		return nil, e
	}

	cmd := cipherscan + " -j --curves -servername " + domain + " " + ip + ":443 "
	fmt.Println(cmd)
	comm := exec.Command("bash", "-c", cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	comm.Stdout = &out
	comm.Stderr = &stderr
	err := comm.Run()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	info := CipherscanOutput{}
	err = json.Unmarshal([]byte(out.String()), &info)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	info.Target = domain
	info.IP = ip

	c, err := info.Stored()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return json.Marshal(c)
}

func getRandomIP(domain string) string {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return ""
	}

	max := len(ips)

	for {
		if max == 0 {
			return ""
		}
		index := rand.Intn(len(ips))

		if ips[index].To4() != nil {
			return ips[index].String()
		} else {
			ips = append(ips[:index], ips[index+1:]...)
		}
		max--
	}
}
