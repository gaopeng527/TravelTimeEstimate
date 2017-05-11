package preprocess
//  处理从OpenStreetMap下载的原始数据，将抽取的数据输出为text文件
import (
	"bufio"
	"fmt"
	"os"
	"encoding/xml"
)

type Node struct {
	Id string `xml:"id,attr"`
	Lat string `xml:"lat,attr"`
	Lon string `xml:"lon,attr"`
}

type Nd struct {
	Ref string `xml:"ref,attr"`
}

type Tag struct {
	K string `xml:"k,attr"`
}

type Way struct {
	Id string 	`xml:"id,attr"`
	Version string 	`xml:"version,attr"`
	Nds []Nd	`xml:"nd"`
	Tags []Tag	`xml:"tag"`
}

type Osm struct {
	Nodes []Node `xml:"node"`
	Ways []Way `xml:"way"`
}

func ParseBigXML(parseFilePath string, storeFilePath string) {
	// 读取文件
	xmlFile, err := os.Open(parseFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	// 创建要存储的文件
	arcFile, err := os.Create(storeFilePath + "\\Arc.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer arcFile.Close()
	arcWriter := bufio.NewWriter(arcFile)
	pointFile, err := os.Create(storeFilePath + "\\Point.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer pointFile.Close()
	pointWriter := bufio.NewWriter(pointFile)

	decoder := xml.NewDecoder(xmlFile)
	var inElement string
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			inElement = se.Name.Local
			// ...and its name is "osm"
			if inElement == "osm" {
				var osm Osm
				// decode a whole chunk of following XML into the
				// variable p which is a Osm (se above)
				decoder.DecodeElement(&osm, &se)

				// Do some stuff with the osm.
				for _, node := range osm.Nodes {
					pointWriter.WriteString(node.Id + "	" + node.Lat + "	" + node.Lon + "\r\n")
					pointWriter.Flush()
					fmt.Println(node.Id + "	" + node.Lat + "	" + node.Lon)
				}
				for _, way := range osm.Ways {
					s := way.Id + "	"+ way.Version
					for _, nd := range way.Nds {
						arcWriter.WriteString(s + "	" + nd.Ref + "\r\n")
						arcWriter.Flush()
						fmt.Println(s + "	" + nd.Ref)
					}
					for _, tag := range way.Tags {
						arcWriter.WriteString(s + "	" + tag.K + "\r\n")
						fmt.Println(s + "	" + tag.K)
					}
				}
			}
		default:
		}

	}
	fmt.Println("文档解析完成！")
}