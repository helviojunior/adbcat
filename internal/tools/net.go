package tools

import (
    "net"
    "encoding/binary"

    "crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"
	"strings"
    "bytes"
    "math/rand"
    "time"
)

type LabeledIP struct {
    IP     net.IP
    Subnet string
}

func ExtractAllSubnets(text string) ([]*net.IPNet, error) {
	ipList := []*net.IPNet{}

    netRe1 := regexp.MustCompile("\\b([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})\\/(3[0-2]|[12][0-9]|[1-9])\\b")
    netRe2 := regexp.MustCompile("\\b([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})\\/(255\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})\\b")
    ipRe := regexp.MustCompile("\\b([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})\\b")

    // Check if is an CIDR Subnet (xxx.xxx.xxx.xxx/xx)
    mlNet1 := netRe1.FindAllStringSubmatch(text, -1)
    for _, mNet1 := range mlNet1 {
	    if len(mNet1) > 0 {
	        _, subnet, err := net.ParseCIDR(mNet1[0])
	        if err != nil {
	            return nil, errors.New("Invalid subnet: " + err.Error())
	        }
	        ipList = append(ipList, subnet)
	    }
	}

    // Check if is an Netmask Subnet (xxx.xxx.xxx.xxx/255.xxx.xxx.xxx)
    mlNet2 := netRe2.FindAllStringSubmatch(text, -1)
    for _, mNet2 := range mlNet2 {
	    if len(mNet2) > 0 {
	        ip := net.ParseIP(mNet2[1])
	        if ip == nil {
	            return nil, errors.New(fmt.Sprintf("Invalid subnet ip (%s)", mNet2[0]))
	        }
	        mask := net.IPMask(net.ParseIP(mNet2[2]).To4())
	        if mask == nil {
	            return nil, errors.New(fmt.Sprintf("Invalid subnet mask (%s)", mNet2[0]))
	        }
	        ones, _ := mask.Size()

	        cidr := fmt.Sprintf("%s/%d", ip.String(), ones)

	        _, subnet, err := net.ParseCIDR(cidr)
	        if err != nil {
	            return nil, errors.New("Invalid subnet: " + err.Error())
	        }
	        ipList = append(ipList, subnet)
	    }
	}

    // Check if is an IP addr
    mlIpRe := ipRe.FindAllStringSubmatch(text, -1)
    for _, mIp := range mlIpRe {
	    if len(mIp) > 0 {
	        ip := net.ParseIP(mIp[0])
	        if ip == nil {
	            return nil, errors.New(fmt.Sprintf("Invalid ip address (%s)", mIp[0]))
	        }
	        _, subnet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ip, 32))
	        if err != nil {
	            return nil, errors.New(fmt.Sprintf("Invalid ip address (%s): %s", mIp[0], err.Error()))
	        }
	        add := true
	        for _, netIp := range ipList {
	        	if netIp.Contains(ip) {
	        		add = false
	        	}
	        }
	        if add {
		        ipList = append(ipList, subnet)
		    }
	    }
	}

    return ipList, nil
}

func IpToUint32(ip net.IP) uint32 {
    ip = ip.To4()
    if ip == nil {
        return 0
    }
    return binary.BigEndian.Uint32(ip)
}

func SubnetToUint32(subNet net.IPNet) uint32 {
	return IpToUint32(subNet.IP)
}

func ParseCertificatePEM(pemData string) (*x509.Certificate, error) {
	pemData = normalizePEM(pemData)
	block, _ := pem.Decode([]byte(pemData))
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Println(pemData)
	    return nil, errors.New("Failed to decode PEM block")
	}

	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
	    return nil, err
	}

	return parsedCert, nil
}

func normalizePEM(input string) string {
	re := regexp.MustCompile(`(?m)^.*\n`)
	txt := re.ReplaceAllStringFunc(input, func(line string) string {
		if strings.HasPrefix(line, "-----") {
			return "\n" + line // keep newline for header/footer
		}
		return strings.TrimSuffix(line, "\n") // remove newline
	})
	return strings.TrimPrefix(txt, "\n")
}

// IsSelfSigned checks if a certificate is self-signed
func IsSelfSigned(cert *x509.Certificate) bool {
    // Check if subject and issuer are equal
    if !bytes.Equal(cert.RawSubject, cert.RawIssuer) {
        return false
    }

    // Try to verify the certificate with its own public key
    err := cert.CheckSignatureFrom(cert)
    return err == nil
}

func SubnetHosts(cidr string) ([]net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		// Copy the IP so we don't overwrite the slice later
		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)
		ips = append(ips, ipCopy)
	}

	// Remove network and broadcast address if applicable
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}


func IpTo24Net(ip net.IP) *net.IPNet {
	if ip.To4() == nil {
		return nil // Not an IPv4 address
	}
	mask := net.CIDRMask(24, 32)
	network := ip.Mask(mask)
	return &net.IPNet{IP: network, Mask: mask}
}

// Expand CIDR into usable IPs and return with its subnet label
func ExpandCIDR(cidr string) ([]LabeledIP, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	var result []LabeledIP
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)
		ones, _ := ipnet.Mask.Size()
		if ones >= 24 {
			result = append(result, LabeledIP{IP: ipCopy, Subnet: ipnet.String()})
		}else{
			result = append(result, LabeledIP{IP: ipCopy, Subnet: IpTo24Net(ip).String()})
		}
	}

	// Remove network/broadcast for IPv4
	/*
	if len(result) > 2 && ip.To4() != nil {
		return result[1 : len(result)-1], nil
	}*/
	return result, nil
}

// Shuffle slice of LabeledIP
func ShuffleLabeledIPs(ips []LabeledIP) {
    rand.Seed(time.Now().UnixNano())
    rand.Shuffle(len(ips), func(i, j int) {
        ips[i], ips[j] = ips[j], ips[i]
    })
}

// Increment an IP address
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}
