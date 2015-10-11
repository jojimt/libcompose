
package nethooks

import (
	"os"
	"os/exec"
	"strings"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

type imageInfo struct {
	portID int
	protoName string
}

func getImageInfo(imageName string) ([]imageInfo, error) {
	imageInfoList := []imageInfo{}

	docker, err := dockerclient.NewDockerClient(os.Getenv("DOCKER_HOST"), nil)
	if err != nil {
		log.Errorf("Unable to connect to docker. Error %v", err)
		return imageInfoList, err
	}

	info, err := docker.InspectImage(imageName)
	log.Debugf("Got the following info for the image %#v", info)

	if err != nil {
		log.Errorf("Unable to inspect image '%s'. Error %v", imageName, err)
		return imageInfoList, err
	}

	for exposedPort := range info.Config.ExposedPorts {
		if strings.Contains(exposedPort, "/") {
			imageInfo := imageInfo{}
			values := strings.Split(exposedPort, "/")
			imageInfo.portID, _ = strconv.Atoi(values[0])
			imageInfo.protoName = values[1]
			log.Infof("Extracted port info %v from image '%s'", imageInfo, imageName)
			imageInfoList = append(imageInfoList, imageInfo)
		}
	}

	return imageInfoList, nil
}

func getContainerIP(contName string) string {
	ipAddress := ""
	output, err := exec.Command("docker", "exec", contName, "/sbin/ip", "address", "show").CombinedOutput()
	if err != nil {
		log.Errorf("Unable to fetch container '%s' IP. Error %v", contName,err)
		return ipAddress
	}

	lines := strings.Split(string(output), "\n")
	for _,line := range lines {
		if strings.Contains(line, "eth0") && strings.Contains(line, "inet ") {
			words := strings.Split(line, " ")
			for idx, word := range words {
				if word == "inet" {
					ipAddress = strings.Split(words[idx+1], "/")[0]
				}
			}
		}
	}

	return ipAddress
}

func populateEtcHosts(contName, dnsSvcName, ipAddress string) error {
	command := "echo " + ipAddress + "     " + dnsSvcName + " >> /etc/hosts"
	if _, err := exec.Command("docker", "exec", contName, "/bin/sh", "-c", command).CombinedOutput(); err != nil {
		log.Errorf("Unable to populate etc hosts. Error %v", err)
		return err
	}

	if output, err := exec.Command("docker", "exec", contName, "cat", "/etc/hosts").CombinedOutput(); err != nil {
		log.Infof("VJ ===> output = %s ", output)
	}
	return nil
}