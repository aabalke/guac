package file

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func GetZipType(path string) string {
	r, err := zip.OpenReader(path)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	for _, f := range r.File {
		switch ext := filepath.Ext(f.Name); ext {
		case ".gb", ".gbc", ".gba", ".nds":
			return ext
		}
	}

	return ""
}

func openFile(f *zip.File) *[]byte {
	r, err := f.Open()
	if err != nil {
		return nil
	}

	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil
	}

	return &b
}

func ReadZip(path string) (name string, rom, sav, rtc *[]byte) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return "", nil, nil, nil
	}
	defer z.Close()

	for _, f := range z.File {

		switch ext := filepath.Ext(f.Name); ext {
		case ".gb", ".gbc", ".gba", ".nds":
			rom = openFile(f)
			name = f.Name
			continue
		}

		switch f.Name {
		case name + ".save":
			sav = openFile(f)

		case name + ".rtc":
			rtc = openFile(f)
		}
	}

	return name, rom, sav, rtc
}

func WriteZip(path, name string, sav *[]byte) {
	writeZip(path, name+".save", *sav)
}

func WriteRtcZip(path, name string, rtc *[]byte) {
	writeZip(path, name+".rtc", *rtc)
}

func writeZip(path, target string, newContent []byte) bool {
	r, err := zip.OpenReader(path)
	if err != nil {
		return false
	}

	tmpPath := path + ".tmp"
	outFile, err := os.Create(tmpPath)
	if err != nil {
		return false
	}

	zw := zip.NewWriter(outFile)

	match := false

	for _, file := range r.File {
		var srcReader io.ReadCloser
		srcReader, err = file.Open()
		if err != nil {
			outFile.Close()
			os.Remove(tmpPath)
			return false
		}

		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   file.Name,
			Method: file.Method,
		})
		if err != nil {
			srcReader.Close()
			outFile.Close()
			os.Remove(tmpPath)
			return false
		}

		if file.Name == target {
			_, err = w.Write(newContent)
			match = true
		} else {
			_, err = io.Copy(w, srcReader)
		}

		srcReader.Close()
		if err != nil {
			outFile.Close()
			os.Remove(tmpPath)
			return false
		}
	}

	if !match {
		w, err := zw.Create(target)
		if err != nil {
			outFile.Close()
			os.Remove(tmpPath)
			return false
		}
		if _, err := w.Write(newContent); err != nil {
			outFile.Close()
			os.Remove(tmpPath)
			return false
		}
	}

	if err := zw.Close(); err != nil {
		outFile.Close()
		os.Remove(tmpPath)
		return false
	}
	if err := outFile.Close(); err != nil {
		os.Remove(tmpPath)
		return false
	}

	if err := r.Close(); err != nil {
		return false
	}

	err = os.Rename(tmpPath, path)
	return err == nil
}
