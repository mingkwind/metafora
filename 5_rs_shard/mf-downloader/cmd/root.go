package cmd

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "metafora-downloader",
	Short: "A tool for uploading and downloading files to METAFORA",
	Run:   start,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	// 参数有网址，上传、下载，文件路径，locate, version,
	rootCmd.Flags().StringP("url", "u", "", "The url of METAFORA server")
	// mode有上传、下载、定位、查看所有版本
	rootCmd.Flags().StringP("mode", "m", "", "The mode of the program, upload, download, locate, version")
	// 如果是上传，需要文件路径
	rootCmd.Flags().StringP("path", "p", "", "The path of the file")
	// 如果是下载，需要文件名
	rootCmd.Flags().StringP("name", "n", "", "The name of the file")
	// 如果是定位，版本号可选
	rootCmd.Flags().StringP("version", "v", "", "The version of the file")
	// url是必须的
	rootCmd.MarkFlagRequired("url")
	// mode是必须的
	rootCmd.MarkFlagRequired("mode")
}

func start(cmd *cobra.Command, args []string) {
	// 首先获取参数
	url, _ := cmd.Flags().GetString("url")
	// 首先判断url是否可连通
	if !checkURL(url) {
		fmt.Println("The url is not valid")
		return
	}
	mode, _ := cmd.Flags().GetString("mode")
	// mode只能是upload、download、locate、version
	switch mode {
	case "upload":
		// 上传
		path, _ := cmd.Flags().GetString("path")
		if path == "" {
			fmt.Println("The path is empty")
			return
		}
		upload(url, path)
	case "download":
		// 下载
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("The name is empty")
			return
		}
		download(url, name)
	case "locate":
		// 定位
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("The name is empty")
			return
		}
		version := 0
		versionStr, _ := cmd.Flags().GetString("version")
		tmpVersion, err := strconv.Atoi(versionStr)
		if err == nil {
			version = tmpVersion
		}
		locate(url, name, version)
	case "version":
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("The name is empty")
			return
		}
		versions(url, name)
	}

}

func checkURL(urlStr string) bool {
	// 判断url是否合法
	parsedUrl, err := url.Parse(urlStr)
	return err == nil && parsedUrl.Scheme != "" && parsedUrl.Host != ""
}

func CalculateHash(r io.Reader) string {
	// 计算sha-256的字符串
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return ""
	}
	// 转化为十六进制字符串
	return hex.EncodeToString(hash.Sum(nil))
}

func upload(urlStr string, path string) {
	// 上传文件
	// 首先将文件进行gzip压缩
	// 然后对压缩包进行hash计算
	// 然后将压缩包上传到服务器
	//将path的最后部分作为文件名
	filename := path[strings.LastIndex(path, "/")+1:]
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Open file failed")
		return
	}
	defer file.Close()
	// 对其进行gzip压缩
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	defer gzWriter.Close()
	if _, err := io.Copy(gzWriter, file); err != nil {
		fmt.Println("Gzip failed")
		return
	}
	gzWriter.Flush()
	// 计算hash
	// buf复制一份
	// 一份用于计算hash
	// 一份用于上传
	gzfile := bytes.NewReader(buf.Bytes())
	hash := CalculateHash(gzfile)
	// 重置gzfile指针位置
	gzfile.Seek(0, 0)
	// 上传文件
	// 在文件头中加入hash
	//使用PUT方法上传文件
	// url合并方法
	uStr, err := url.JoinPath(urlStr, "objects", filename)
	if err != nil {
		fmt.Println("url join failed")
	}
	request, _ := http.NewRequest("PUT", uStr, gzfile)
	request.Header.Add("Digest", "SHA-256="+hash)
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Upload failed")
		return
	}
	// 打印返回的信息
	// 读取返回的信息
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Read body failed")
		return
	}
	fmt.Println(string(body))
}

func download(urlStr, name string) {
	// 下载文件
	// 首先获取文件的hash
	// 然后下载文件
	// 然后对文件进行hash计算
	// 然后对比hash
	// 然后解压文件
	// 然后将文件保存
	// 从服务器获取文件的hash
	uStr, err := url.JoinPath(urlStr, "objects", name)
	if err != nil {
		fmt.Println("url join failed")
	}
	response, err := http.Get(uStr)
	if err != nil {
		fmt.Println("Get file failed")
		return
	}
	// 获取文件的hash// 等号后面的是hash值
	hash := response.Header.Get("Digest")[8:]
	// 下载文件
	response, err = http.Get(uStr)
	if err != nil {
		fmt.Println("Get file failed")
		return
	}
	// 一边计算hash，一边写入到buf中
	// 创建一个缓冲区
	var buf bytes.Buffer
	defer buf.Reset()
	hashReader := io.TeeReader(response.Body, &buf)
	// 计算hash
	newHash := CalculateHash(hashReader)
	// 判断hash是否相同
	if hash != newHash {
		fmt.Println("The file is not complete")
		return
	} else {
		// 创建gzip解压读取器
		gzReader, _ := gzip.NewReader(&buf)
		defer gzReader.Close()
		//解压文件后将其保存为name
		newFile, _ := os.Create(name)
		defer newFile.Close()
		// 将解压后的文件写入新文件中
		io.Copy(newFile, gzReader)
	}
}

func locate(urlStr, name string, version int) {
	// 从服务器获取文件的hash
	uStr, err := url.JoinPath(urlStr, "locate", name)
	if err != nil {
		fmt.Println("url join failed")
	}
	if version != 0 {
		uStr += "?version=" + strconv.Itoa(version)
	}
	response, err := http.Get(uStr)
	if err != nil {
		fmt.Println("Get file failed")
		return
	}
	// 打印返回的信息
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Read body failed")
		return
	}
	fmt.Println(string(body))
}

func versions(urlstr, name string) {
	// 获取服务器上的版本号
	uStr, err := url.JoinPath(urlstr, "version", name)
	if err != nil {
		fmt.Println("url join failed")
	}
	response, err := http.Get(uStr)
	if err != nil {
		fmt.Println("Get file failed")
		return
	}
	// 打印返回的信息
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Read body failed")
		return
	}
	fmt.Println(string(body))
}
