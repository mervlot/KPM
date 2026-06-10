package install

import (
	"fmt"
	"kpm/types"
	"kpm/utils"
	"testing"
)
	var mave types.Mavenurl = types.Mavenurl{
		Group: "io.kvision",
		Artifact: "kvisionserver-javalin",
		Version: "1.2.0",

	}

func TestXmlParsing(t *testing.T) {
 if utils.IsFile(mave.LocalMavenMetaData()) {
		fmt.Println("it exists")
	}else{
		fmt.Println("not found")

	}
}
