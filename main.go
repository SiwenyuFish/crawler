package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// 定义比赛数据的结构
type MatchData struct {
	Time  string `json:"time"`
	Sport string `json:"sport"`
	Name  string `json:"name"`
	Venue string `json:"venue"`
}

// MedalData 定义奖牌数据的结构
type MedalData struct {
	Rank    string `json:"rank"`
	Country string `json:"country"`
	Gold    string `json:"gold"`
	Silver  string `json:"silver"`
	Bronze  string `json:"bronze"`
	Total   string `json:"total"`
}

func main() {
	// 爬取奖牌数据
	if err := fetchMedalData(); err != nil {
		log.Fatal(err)
	}

	// 爬取比赛数据
	if err := fetchMatchData(); err != nil {
		log.Fatal(err)
	}
}

// 爬取并保存奖牌数据
func fetchMedalData() error {
	// 创建Chromedp上下文
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 设置超时上下文
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 目标网页
	url := "https://sports.cctv.cn/Paris2024/medal_list/index.shtml?spm=C73465.PkN5JcjBF6mp.E6mpRwlrGbbT.1"

	var htmlContent string

	// 启动浏览器并抓取页面的HTML内容
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("#medal_list1"),
		chromedp.OuterHTML("body", &htmlContent),
	)
	if err != nil {
		return err
	}

	// 解析抓取到的HTML并提取数据
	medalData := extractMedalData(htmlContent)

	// 将奖牌数据转换为JSON格式并保存到文件
	return saveToJSON(medalData, "medal_data.json")
}

// 提取奖牌数据
func extractMedalData(htmlContent string) []MedalData {
	var medalList []MedalData

	// 解析HTML内容
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	// 查找表格中的每一行数据
	doc.Find("#medal_list1 tr").Each(func(i int, row *goquery.Selection) {
		rank := row.Find("td").Eq(0).Text()
		countryLink, _ := row.Find("td.country a").Attr("href") // 提取链接
		countryID := extractCountryID(countryLink)              // 从链接中提取 countryid
		gold := row.Find("td").Eq(2).Text()
		silver := row.Find("td").Eq(3).Text()
		bronze := row.Find("td").Eq(4).Text()
		total := row.Find("td").Eq(5).Text()

		// 清理数据
		rank = strings.TrimSpace(rank)
		countryID = strings.TrimSpace(countryID)
		gold = strings.TrimSpace(gold)
		silver = strings.TrimSpace(silver)
		bronze = strings.TrimSpace(bronze)
		total = strings.TrimSpace(total)

		// 创建MedalData对象并添加到列表
		if rank != "" {
			medalData := MedalData{
				Rank:    rank,
				Country: countryID, // 使用提取到的 countryid
				Gold:    gold,
				Silver:  silver,
				Bronze:  bronze,
				Total:   total,
			}
			medalList = append(medalList, medalData)
		}
	})

	return medalList
}

// 从链接中提取 countryid
func extractCountryID(link string) string {
	// 查找 "countryid=" 后面的值
	parts := strings.Split(link, "countryid=")
	if len(parts) > 1 {
		return strings.Split(parts[1], "&")[0]
	}
	return ""
}

// 爬取并保存比赛数据
func fetchMatchData() error {
	startDate := "2024-07-24"
	endDate := "2024-08-11"

	// 解析日期
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return err
	}

	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		dateStr := current.Format("20060102") // 格式化为 yyyyMMdd
		fileName := fmt.Sprintf("%s_data.json", dateStr)
		url := fmt.Sprintf("https://sports.cctv.cn/Paris2024/schedule/date/index.shtml?date=%s", dateStr)

		// 创建Chromedp上下文
		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()

		// 设置超时上下文
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		var htmlContent string

		// 启动浏览器并抓取页面的HTML内容
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitReady("#data_list"),
			chromedp.OuterHTML("body", &htmlContent),
		)
		if err != nil {
			log.Printf("Failed to fetch data for date %s: %v", dateStr, err)
			continue // 失败时继续下一天的抓取
		}

		// 解析抓取到的HTML并提取数据
		matchData := extractMatchData(htmlContent)

		// 将比赛数据转换为JSON格式并保存到文件
		err = saveToJSON(matchData, fileName)
		if err != nil {
			log.Printf("Failed to save data for date %s: %v", dateStr, err)
			continue // 失败时继续下一天的保存
		}

		fmt.Printf("数据已成功保存到 %s 文件中\n", fileName)
	}

	return nil
}

// 提取比赛数据
func extractMatchData(htmlContent string) []MatchData {
	var matchList []MatchData

	// 解析HTML内容
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	// 查找表格中的每一行数据
	doc.Find("#data_list tr").Each(func(i int, row *goquery.Selection) {
		time := row.Find("td").Eq(0).Text()
		sport := row.Find("td").Eq(2).Text()
		name := row.Find("td").Eq(3).Text()
		venue := row.Find("td").Eq(4).Text()

		// 清理数据
		time = strings.TrimSpace(time)
		sport = strings.TrimSpace(sport)
		name = strings.TrimSpace(name)
		venue = strings.TrimSpace(venue)

		// 创建MatchData对象并添加到列表
		if time != "" {
			matchData := MatchData{
				Time:  time,
				Sport: sport,
				Name:  name,
				Venue: venue,
			}
			matchList = append(matchList, matchData)
		}
	})

	return matchList
}

// 将数据保存为JSON文件
func saveToJSON(data interface{}, filename string) error {
	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将数据编码为JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 格式化输出

	return encoder.Encode(data)
}
