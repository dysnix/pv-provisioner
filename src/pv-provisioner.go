package main

import (
	"flag"
	"fmt"
	"github.com/dysnix/pv-provisioner/src/pkg/amazon"
	"github.com/dysnix/pv-provisioner/src/pkg/gcp"
)

const CloudNameAWS = "aws"
const CloudNameGCP = "gcp"

func main() {
	cloud := flag.String("cloud", CloudNameGCP, "select specific cloud provider")
	flag.Parse()

	fmt.Println("Selected cloud provider:", *cloud)

	switch *cloud {
	case CloudNameGCP:
		gcp.RunProcessor()
	case CloudNameAWS:
		amazon.RunProcessor()
	default:
		fmt.Println("Unsupported cloud provider:", *cloud)
	}
}
