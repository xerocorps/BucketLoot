package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"math/rand"
	"net"
	"log"
)

var dnsServers = []string{
	"8.8.8.8",   // Google Public DNS (IPv4)
	"1.1.1.1",   // Cloudflare DNS (IPv4)
	"208.67.222.222", // OpenDNS (IPv4)
	"9.9.9.9",   // Quad9 DNS (IPv4)
	"75.75.75.75", // Comcast DNS (IPv4)
	"2001:4860:4860::8888", // Google Public DNS (IPv6)
	"2606:4700:4700::1111", // Cloudflare DNS (IPv6)
	"2620:0:ccc::2", // OpenDNS (IPv6)
	"2620:fe::9",    // Quad9 DNS (IPv6)
	"2001:558:feed::1", // Comcast DNS (IPv6)
	"209.244.0.3",
	"209.244.0.4",
	"8.8.4.4",
	"8.26.56.26",
	"8.20.247.20",
	"208.67.222.222",
	"208.67.220.220",
	"156.154.70.1",
	"156.154.71.1",
	"199.85.126.10",
	"199.85.127.10",
	"81.218.119.11",
	"209.88.198.133",
	"195.46.39.39",
	"195.46.39.40",
	"216.87.84.211",
	"23.90.4.6",
	"199.5.157.131",
	"208.71.35.137",
	"208.76.50.50",
	"208.76.51.51",
	"216.146.35.35",
	"216.146.36.36",
	"89.233.43.71",
	"89.104.194.142",
	"74.82.42.42",
	"109.69.8.51",
}

func main() {
	// Check for network connectivity
	for {
		if isInternetConnected() {
			break
		}
		log.Println("Waiting for internet connection...")
		time.Sleep(30 * time.Second) // Wait for 30 seconds before rechecking
	}
	takeInput()
	fmt.Println(banner, "")
	fmt.Println("\n ")
	if len(args) > 0 {
		fmt.Println("Processing arguments...")

		for i := 0; i < len(args); i++ {
			arg := args[i]
			if strings.HasSuffix(arg, ".txt") {
				fmt.Println("Reading file content from " + arg + "...")
				readFile(arg)
			} else if urlValidation.MatchString(arg) {
				allURLs = append(allURLs, arg)
			} else if arg == "slow" || arg == "-slow" || arg == "--slow" {
				*slowScan = true
			} else if arg == "dig" || arg == "-dig" || arg == "--dig" {
				*digMode = true
			} else if arg == "notify" || arg == "-notify" || arg == "--notify" {
				*notify = true
			} else if arg == "log-errors" || arg == "-log-errors" || arg == "--log-errors" {
				*errorLogging = true
			} else if arg == "full" || arg == "--full" || arg == "-full" {
				readCredsFile()
			} else if arg == "max-size" || arg == "-max-size" || arg == "--max-size" {
				if i+1 < len(args) {
					maxSizeStr := args[i+1]
					_, err := strconv.Atoi(maxSizeStr)
					if err != nil {
						log.Fatalln("Invalid max size:", maxSizeStr, ".. [Exiting!]")
					} else {
						maxFileSize = maxSizeStr
					}
				} else {
					log.Fatalln("Missing max size argument.. [Exiting!]")
				}
				i++ // Skip the next argument since it has been processed
			} else if arg == "search" || arg == "-search" || arg == "--search" {
				if i+1 < len(args) {
					if strings.Contains(args[i+1], ":::") {
						keywords := strings.Split(args[i+1], ":::")
						for _, keyword := range keywords {
							if strings.HasSuffix(keyword, ".txt") {
								file, err := os.Open(keyword)
								if err != nil {
									log.Fatalln("[Error] Looks like the tool is facing some issue while loading the specified file. [", err.Error(), "]")
								}
								defer file.Close()

								scanner := bufio.NewScanner(file)
								for scanner.Scan() {
									scanKeywords = append(scanKeywords, scanner.Text())
								}
							} else {
								scanKeywords = append(scanKeywords, keyword)
							}
						}
					} else {
						if strings.HasSuffix(args[i+1], ".txt") {
							file, err := os.Open(args[i+1])
							if err != nil {
								log.Fatalln("[Error] Looks like the tool is facing some issue while loading the specified file. [", err.Error(), "]")
							}
							defer file.Close()

							scanner := bufio.NewScanner(file)
							for scanner.Scan() {
								scanKeywords = append(scanKeywords, scanner.Text())
							}
						} else {
							scanKeywords = append(scanKeywords, args[i+1])
						}
					}
				} else {
					log.Fatalln("Missing search argument.. [Exiting!]")
				}
				i++ // Skip the next argument since it has been processed
			} else if arg == "save" || arg == "-save" || arg == "--save" {
				if i+1 < len(args) {
					if strings.HasSuffix(args[i+1], ".txt") || strings.HasSuffix(args[i+1], ".json") {
						saveOutput = args[i+1]
					} else {
						saveOutput = "output.json"
					}
				} else {
					saveOutput = "output.json"
				}
				i++ // Skip the next argument since it has been processed
			}
		}
		if *notify {
			notifyErr := loadNotifyConfig()
			if notifyErr != nil {
				fmt.Println("Looks like these is some issue with your notifyconfig file:", notifyErr)
				toJSON()
				os.Exit(1)
			}
		}
		allURLs = formatURL(allURLs)
		if len(allURLs) > 0 {
			listS3BucketFiles(allURLs)
			if len(iniFileListData.Scannable) > 0 {
				if len(iniFileListData.NotScannable) > 0 {
					bucketlootOutput.Skipped = iniFileListData.NotScannable
				}
				bucketlootOutput.Scanned = iniFileListData.Scannable
				fmt.Println("\n ")
				fmt.Println("Discovered a total of " + strconv.Itoa(iniFileListData.TotalFiles) + " bucket files...")
				fmt.Println("Total bucket files of interest: " + strconv.Itoa(iniFileListData.TotalIntFiles))
				fmt.Println("\n ")
				if *slowScan {
					fmt.Println("Starting to scan the files... [SLOW]")
				} else {
					fmt.Println("Starting to scan the files... [FAST]")
				}
				for _, bucketEntry := range iniFileListData.ScanData {
					if *slowScan {
						scanS3FilesSlow(bucketEntry.IntFiles, bucketEntry.URL)
					} else {
						scanS3FilesFast(bucketEntry.IntFiles, bucketEntry.URL)
					}
				}
				toJSON()
			} else {
				fmt.Println("Oops.. Looks like no interesting buckets were discovered! Aborting the scan...")
				bucketlootOutput.Skipped = iniFileListData.NotScannable
				toJSON()
			}
		} else {
			log.Fatalln("Looks like no valid URLs/domains were specified.. [Exiting!]")
		}
	} else {
		fmt.Println("Looks like no arguments were specified.. [Exiting!]")
	}
}

func isInternetConnected() bool {
	rand.Seed(time.Now().UnixNano())

	// Shuffle the dnsServers slice randomly
	shuffledDNS := make([]string, len(dnsServers))
	copy(shuffledDNS, dnsServers)
	rand.Shuffle(len(shuffledDNS), func(i, j int) {
		shuffledDNS[i], shuffledDNS[j] = shuffledDNS[j], shuffledDNS[i]
	})
	// Check if there's an active network connection
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println("[INTERNET CHECK ERROR]:", err)
		return false
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				for {
					for _, dnsServer := range shuffledDNS {
						_, err := net.LookupHost(dnsServer)
						if err == nil {
							return true
						}
					}
			
					log.Println("Waiting for internet connection...")
					time.Sleep(2 * time.Second) // Wait for 5 seconds before rechecking
				}
			}
		}
	}
	return false
}
