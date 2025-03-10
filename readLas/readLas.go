package readLas

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

type LAS struct {
	Info   GeomInfo
	Points []Point
}

func Read_file(filePath string) LAS {
	// Open and Read Binary File
	file, errFile := os.Open(filePath)
	stat, errStat := os.Stat(filePath)
	defer file.Close()
	if errFile != nil {
		log.Fatal(errFile)
	}
	if errStat != nil {
		log.Fatal(errStat)
	}

	fileLength := stat.Size()
	bytesBuffer := make([]byte, fileLength)
	bin, err := file.Read(bytesBuffer)
	if err != nil {
		log.Fatal(err)
	}
	// Store to array byte
	var data []byte = bytesBuffer[:bin]

	//parsing header
	var offsetToPoint uint32 = binary.LittleEndian.Uint32(data[96:100])
	var geoInfo GeomInfo = ReadGeomInfo(data[131:143])
	var arrPoint = make([]Point, int((int(fileLength)-int(offsetToPoint))/29))

	//assign point to array
	for i := 0; i < int((int(fileLength)-int(offsetToPoint))/29); i++ {
		arrPoint[i] = ReadPoint(data[int(offsetToPoint)+29*i : (int(offsetToPoint) + 29*(i+1))])
	}
	var result LAS
	result.Info = geoInfo
	result.Points = arrPoint
	return result
}

type GeomInfo struct {
	Scale  CoordXYZ
	Offset CoordXYZ
	Extent CoordExtent
}
type CoordExtent struct {
	MaxX float64
	MinX float64
	MaxY float64
	MinY float64
	MaxZ float64
	MinZ float64
}
type CoordXYZ struct {
	X float64
	Y float64
	Z float64
}

func ReadGeomInfo(bin []byte) GeomInfo {
	var info GeomInfo
	var _scale CoordXYZ
	var _offset CoordXYZ
	var _extent CoordExtent
	binary.Read(bytes.NewReader(bin[0:24]), binary.LittleEndian, &_scale)
	binary.Read(bytes.NewReader(bin[24:48]), binary.LittleEndian, &_offset)
	binary.Read(bytes.NewReader(bin[48:48+24*2]), binary.LittleEndian, &_extent)

	info.Scale = _scale
	info.Offset = _offset
	info.Extent = _extent
	return info
}

type Point struct {
	X              int32
	Y              int32
	Z              int32
	Intensity      uint16
	Classification uint8
	Red            uint16
	Green          uint16
	Blue           uint16
}

func ReadPoint(bin []byte) Point {
	var point Point

	binary.Read(bytes.NewReader(bin[0:4]), binary.LittleEndian, &point.X)
	binary.Read(bytes.NewReader(bin[4:8]), binary.LittleEndian, &point.Y)
	binary.Read(bytes.NewReader(bin[8:12]), binary.LittleEndian, &point.Z)
	binary.Read(bytes.NewReader(bin[12:14]), binary.LittleEndian, &point.Intensity)
	binary.Read(bytes.NewReader(bin[19:20]), binary.LittleEndian, &point.Classification)
	binary.Read(bytes.NewReader(bin[21:23]), binary.BigEndian, &point.Red)
	binary.Read(bytes.NewReader(bin[23:25]), binary.BigEndian, &point.Green)
	binary.Read(bytes.NewReader(bin[25:27]), binary.BigEndian, &point.Blue)

	return point
}
