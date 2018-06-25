package main

import (
	"github.com/fogleman/gg"
	"fmt"
	"os"
	"image/png"
	"image/color"
	"io/ioutil"
	"runtime"
	"strings"
	"sort"
)

const(
	transparent = 1
	notTransparent = 0
)

type WorldMap struct{
	Width int
	Height int
	CollisionMap []int
	RealImageWidth int
	RealImageHeight int
}

type Position struct{
	Xpos int
	Ypos int
	Value int
}

type WordCount struct{
	word string
	count int
}

type WordCountList []*WordCount

func main(){
	//初始化原始图片, 字体, 最大字号
	fontpath := "fz.ttf"
	wm := TwoByBitmap("tiger.png")
	fontsize := 60

	//初始化文本列表
	textlist := []string{"Apple", "天气", "今天", "多云"}


	//dc为写入文本的context, mc为测量文本区域的context
	dc := gg.NewContext(wm.RealImageWidth, wm.RealImageHeight)
	mc := gg.NewContext(wm.RealImageWidth, wm.RealImageHeight)
	//初始化
	dc.SetRGBA(0,0,0,0)
	dc.Clear()
	dc.SetRGB(0,0,0)
	c := 0
	for {
		//分别进行字体字号初始化
		if err := dc.LoadFontFace(fontpath, float64(fontsize)); err != nil {
			panic(err)
		}
		if err := mc.LoadFontFace(fontpath, float64(fontsize)); err != nil {
			panic(err)
		}
		colour := selectcolor(c)
		dc.SetRGB(float64(colour.R), float64(colour.G), float64(colour.B))
		//遍历文本
		for _,text := range textlist{
			mc.SetRGBA(0, 0, 0, 0)
			mc.Clear()
			mc.SetRGB(0, 0, 0)
			//查找文本边界, 进行碰撞检测, 如果找到位置, 将文本写入dc, 并更新原图的Bitmap
			textW, textH, textpos := GetTextBound(mc, text)
			fx, fy, isfound := checkCP(wm, textW, textH, textpos)
			if isfound{
				dc.DrawStringAnchored(text, float64(fx), float64(fy), 0.5, 0.5)
				wm.updateWM(fx-textW/2,fy-textH/2,textW, textH, textpos)
				continue
			}
			//如果水平方向的文本找不到位置, 尝试垂直的文本
			wm.rotateWM()
			textW, textH, textpos = GetTextBound(mc, text)
			fx, fy, isvfound := checkCP(wm, textW, textH, textpos)
			if isvfound{
				//以旋转画布的方式写入垂直的文本
				dc.RotateAbout(gg.Radians(float64(90)), float64(wm.RealImageWidth/2), float64(wm.RealImageHeight/2))
				dc.DrawStringAnchored(text, float64(fx), float64(fy), 0.5, 0.5)
				dc.RotateAbout(gg.Radians(float64(-90)), float64(wm.RealImageWidth/2), float64(wm.RealImageHeight/2))
				wm.updateWM(fx-textW/2,fy-textH/2,textW, textH, textpos)
				wm.rotateWMB()
				continue
			}else{
				wm.rotateWMB()
			}
		}
		c++
		if c > 2{
			c = 0
		}
		//字号处理, 大字号应该较少, 小字号应该较多
		fontsize -= 3
		if fontsize < 5{
			break
		}
	}

	/*if err := dc.LoadFontFace(fontpath, float64(fontsize)); err != nil {
		panic(err)
	}
	if err := mc.LoadFontFace(fontpath, float64(fontsize)); err != nil {
		panic(err)
	}
	colour := selectcolor(1)
	dc.SetRGB(float64(colour.R), float64(colour.G), float64(colour.B))
	mc.SetRGBA(0, 0, 0, 0)
	mc.Clear()
	mc.SetRGB(0, 0, 0)
	textW, textH, textpos := GetTextBound(mc, textlist[0])
	fx, fy, checkfound := check(wm, textW, textH, textpos)
	if checkfound {
		dc.DrawStringAnchored(textlist[0], float64(fx), float64(fy), 0.5, 0.5)
		wm.updateWM(fx-textW/2,fy-textH/2,textW, textH, textpos)
	}
	wm.rotateWM()
	textW, textH, textpos = GetTextBound(mc, textlist[1])
	fx, fy, checkfound = check(wm, textW, textH, textpos)
	if checkfound{
		dc.RotateAbout(gg.Radians(float64(90)), 375, 375)
		dc.DrawStringAnchored(textlist[1], float64(fx), float64(fy), 0.5, 0.5)
		dc.RotateAbout(gg.Radians(float64(-90)), 375, 375)
		wm.rotateWMB()
		wm.updateWM(fx-textW/2,fy-textH/2,textW, textH, textpos)
	}else{
		wm.rotateWMB()
	}*/
	//保存画布
	dc.SavePNG("rgbn.png")
	test()
}


//把图片转换成Bitmap, 只分透明的点和不透明的点
func TwoByBitmap(imgpath string) *WorldMap{
	wm := &WorldMap{CollisionMap:make([]int, 0)}
	file,_ := os.Open(imgpath)
	img,_ := png.Decode(file)
	file.Close()
	bounds := img.Bounds()
	w, h := bounds.Size().X, bounds.Size().Y
	wm.RealImageWidth = w
	wm.RealImageHeight = h
	fmt.Println(img.At(0,0))
	//遍历所有点, 只记录alpha值, 也可提取色彩
	for y:=0;y<h;y++{
		for x:=0;x<w;x++{
			colour := img.At(x,y)
			_,_,_,alpha := colour.RGBA()
			//alpha==0时透明
			if alpha == 0{
				wm.CollisionMap = append(wm.CollisionMap, transparent)
			}else{
				wm.CollisionMap = append(wm.CollisionMap, notTransparent)
			}
		}
	}
	return wm
}


//获得文本的边界, 以及文本中实际笔划的占位
func GetTextBound(dc *gg.Context, text string) (int, int, []*Position){
	//初始化
	positions := make([]*Position, 0)
	dc.SetRGBA(0,0,0,0)
	dc.Clear()
	dc.SetRGB(0,0,0)
	//将相应字号的文本画在画布的中间
	//dc.DrawStringAnchored(text, 375, 375, 0.5, 0.5)
	img := dc.Image()
	width, height := dc.Width(), dc.Height()
	dc.DrawStringAnchored(text, float64(width/2), float64(height/2), 0.5, 0.5)
	//查找边界, 文本范围的最小x,y和最大x,y
	maxX, maxY, minX, minY := 0, 0, 9999, 9999
	for y:=0;y<height;y++{
		for x:=0;x<width;x++{
			colour := img.At(x,y)
			_,_,_,alpha := colour.RGBA()
			if alpha != 0{
				if minX > x{
					minX = x
				}
				if minY > y{
					minY = y
				}
				if maxX < x{
					maxX = x
				}
				if maxY < y{
					maxY = y
				}
			}
		}
	}
	//比较某字号下文本的实际边界与理论边界, 取较大的值
	orgW, orgH := dc.MeasureString(text)
	realW, realH := maxX - minX, maxY - minY
	if realW < int(orgW){
		realW = int(orgW)
	}
	if realH < int(orgH){
		realH = int(orgH)
	}
	fmt.Println("real:", realW, realH)
	//记录文本实际笔划的占位, 透明及不透明
	for y:=0;y<realH;y++{
		for x:=0;x<realW;x++{
			colour := img.At(minX+x, minY+y)
			_,_,_,alpha := colour.RGBA()
			if alpha != 0{
				pos := &Position{x,y,notTransparent}
				positions = append(positions, pos)
			}else{
				pos := &Position{x,y,transparent}
				positions = append(positions, pos)
			}
		}
	}
	return realW, realH, positions
}

/*func check(wm *WorldMap, textwidth, textheight int , positions []*Position) (int, int, bool){
	//isFound := true
	var isFound bool
	finalX, finalY := 0, 0
	checkX, checkY := 0, 0
	n := 0
	for{
		if checkX + textwidth > wm.RealImageWidth && checkY + textheight > wm.RealImageHeight{
			isFound = false
			break
		}
		isfit := checkfit(checkX, checkY, wm, textwidth, textheight, positions)
		n++
		if isfit{
			finalX, finalY = checkX+textwidth/2, checkY+textheight/2
			fmt.Println(checkX, checkY, finalX, finalY)
			//wm.updateWM(checkX, checkY, textwidth, textheight, positions)
			isFound = true
			break
		}
		checkX++
		if checkX + textwidth > wm.RealImageWidth{
			checkY++
			checkX = 0
		}
	}
	return finalX, finalY, isFound
}*/


//判断以checkX, checkY开头的文本区域, 是否与原图透明区域发生碰撞
func checkfit(checkX, checkY int, wm *WorldMap, textwidth, textheight int, positions  []*Position) bool{
	for x:=0;x<textwidth;x++{
		for y:=0;y<textheight;y++ {
			//lindex用在原图区域, sindex用在文本区域, 当原图区域为透明, 而文本区域为不透明时, 返回不合适
			lindex := (y+checkY)*wm.RealImageWidth + x + checkX
			sindex := y*textwidth + x
			if positions[sindex].Value == notTransparent{
				if wm.CollisionMap[lindex] == transparent {
					return false
				}
			}
		}
	}
	fmt.Println(wm.CollisionMap[0])
	return true
}

//当文本找到合适的位置后, 更新原图的Bitmap
func (wm *WorldMap) updateWM(pointX, pointY, textwidth, textheight int, positions []*Position){
	//以pointX,pointY开头的文本区域在原图中的投影, 并进行透明度覆盖, 应该用这个方法更新, 但实际效果不太好, 所以改用下面更简单粗暴的方法
	/*for y:=0;y<textheight;y++{
		for x:=0;x<textwidth;x++{
			sindex := y*textwidth + x
			lindex := (y+pointY)*wm.RealImageWidth + x + pointX
			if positions[sindex].Value == notTransparent &&
				wm.CollisionMap[lindex] == notTransparent{
				wm.CollisionMap[lindex] = transparent
			}
		}
	}*/
	//以pointX,pointY开头的文本区域在原图中的投影, 此处-3,3可调, 减少不必要的碰撞
	for y:=-3;y<textheight+3;y++{
		for x:=-3;x<textwidth+3;x++{
			lindex := (y+pointY)*wm.RealImageWidth + x + pointX
			wm.CollisionMap[lindex] = transparent
		}
	}
}

//选择文本颜色
func selectcolor(i int) *color.RGBA{
	colors := []*color.RGBA{
		&color.RGBA{0x0, 0x60, 0x30, 0xff},
		&color.RGBA{0x60, 0x0, 0x0, 0xff},
		&color.RGBA{0x73, 0x73, 0x0, 0xff},
	}
	return colors[i]
}

//检测是否可写入某字号下的文本
//func getCP(wm *WorldMap, textwidth, textheight int){
func checkCP(wm *WorldMap, textwidth, textheight int, positions []*Position) (int, int, bool){
	//textwidth, textheight, textpos := GetTextBound(dc, text)
	//遍历所有文本中心点之前, 先计算原图可写入范围的边界, 如果原图是长方形, 取边界较大的值, 之后会处理不在范围内的情况
	centerX, centerY := wm.RealImageWidth/2, wm.RealImageHeight/2
	edge := 0
	finalX, finalY := 0, 0
	isFound := false
	if wm.RealImageWidth - textwidth >= wm.RealImageHeight - textheight{
		edge = (wm.RealImageWidth - textwidth) / 2
	}else{
		edge = (wm.RealImageHeight - textheight) / 2
	}
	n:=0
	//遍历所有文本中心点
	AllFrame:
	for i:=0;i<edge;i++{
		n++
		//centerPoints := make([][]int, 0)
		//从原图中心点开始, 逐渐往外扩展同心正方形, 所有可能的文本中心点都在这些同心正方形的边上
		minX, maxX := centerX - i, centerX + i
		minY, maxY := centerY - i, centerY + i
		if minX < textwidth/2{
			minX = textwidth/2
		}
		if maxX > wm.RealImageWidth - textwidth/2{
			maxX = wm.RealImageWidth - textwidth/2
		}
		if minY < textheight/2{
			minY = textheight/2
		}
		if maxY > wm.RealImageHeight - textheight/2{
			maxY = wm.RealImageHeight - textheight/2
		}
		//遍历每个正方形的边
		for x:=minX;x<maxX;x++{
			for y:=minY;y<maxY;y++{
				//如果原图为长方形, 直接略过可能在原图范围外的正方形的边
				if y!=minY && y!=maxY && x!=minX && x!=maxX{
					continue
				}else{
					//centerPoint := make([]int, 0)
					//centerPoint = append(centerPoint, x)
					//centerPoint = append(centerPoint, y)
					//centerPoints = append(centerPoints, centerPoint)
					//isfit := checkfit(centerPoint[0]-textwidth/2, centerPoint[1]-textheight/2, wm, textwidth, textheight)
					//用文本中心点计算文本的初始点
					checkX, checkY := x - textwidth/2, y - textheight/2
					isfit := checkfit(checkX, checkY, wm, textwidth, textheight, positions)
					if isfit{
						finalX, finalY = x, y
						fmt.Println(finalX, finalY)
						//wm.updateWM(checkX, checkY, textwidth, textheight, textpos)
						isFound = true
						break AllFrame
					}
/*					isvfit := checkfitv(checkX, checkY, wm, textwidth, textheight)
					if isvfit{
						finalX, finalY = x, y
						fmt.Println(finalX, finalY)
						wm.updateWM(checkX, checkY, textwidth, textheight, textpos)
						isFound = true
						isVeritical = true
						break AllFrame
					}*/
				}
			}
		}
	}
	fmt.Println("centerpoint:", n)
	return finalX, finalY, isFound
}


//90度旋转原图
func (wm *WorldMap) rotateWM(){
	newCMap := make([]int, 0)
	//长宽交换
	wm.RealImageWidth, wm.RealImageHeight = wm.RealImageHeight, wm.RealImageWidth
	//依次查找旋转后的点, 并重新排序
	for x:=wm.RealImageHeight-1;x>=0;x--{
		for y:=0;y<wm.RealImageWidth;y++{
			index := y * wm.RealImageHeight + x
			position := wm.CollisionMap[index]
			newCMap = append(newCMap, position)
		}
	}
	wm.CollisionMap = newCMap
}

//将原图旋转回去
func (wm *WorldMap) rotateWMB(){
	newCMap := make([]int, 0)
	wm.RealImageWidth, wm.RealImageHeight = wm.RealImageHeight, wm.RealImageWidth
	for x:=0;x<wm.RealImageHeight;x++{
		for y:=wm.RealImageWidth-1;y>=0;y--{
			index := y * wm.RealImageHeight + x
			position := wm.CollisionMap[index]
			newCMap = append(newCMap, position)
		}
	}
	wm.CollisionMap = newCMap
}


//test
func test(){
	const S = 1024
	dc := gg.NewContext(S, S)
	dc.SetRGBA(0, 0, 0, 0.1)
	//dc.Push()
	dc.RotateAbout(gg.Radians(float64(15)), S/2, S/2)
	dc.DrawEllipse(S/2, S/2, S*7/16, S/8)
	dc.Fill()
	//dc.Pop()
	dc.SavePNG("rotate1.png")
}


//下面为读取文档词频
func (list WordCountList) Len() int{
	return len(list)
}

func (list WordCountList) Less(i,j int) bool{
	if list[i].count > list[j].count{
		return true
	}else if list[i].count < list[j].count{
		return false
	}else{
		return list[i].word < list[j].word
	}
}

func (list WordCountList) Swap(i,j int) {
	list[i], list[j] = list[j], list[i]
}

func (list WordCountList) Counts() int{
	counts := 0
	for _, v := range list{
		counts += v.count
	}
	return counts
}

func checkCN(textfile string) bool{
	data,_ := ioutil.ReadFile()
	i := 0
	for {
		if string(data[i]) != " "{
			if data[i] > 128{
				return true
			}else{
				return false
			}
		}
		i++
		if i > 1000{
			break
		}
	}
	return false
}

func getTextList(textfile string) WordCountList{
	data,err := ioutil.ReadFile(textfile)
	if os.IsNotExist(err){
		fmt.Println(err)
	}
	text := string(data)
	newRountineCount := runtime.NumCPU() * 2 - 1
	runtime.GOMAXPROCS(newRountineCount + 1)
	parts := splitText(text, newRountineCount)
	ch := make(chan map[string]int, newRountineCount)
	for i:=0;i<newRountineCount;i++{
		go getWordCount(parts[i], ch)
	}

	wordMap := make(map[string]int, 0)
	done := 0
	for{
		received := <- ch
		done ++
		for k,v := range received{
			wordMap[strings.ToLower(k)] += v
		}
		if done == newRountineCount{
			break
		}
	}

	list := make(WordCountList, 0)
	for k,v := range wordMap{
		wordcount := &WordCount{k, v}
		list = append(list, wordcount)
	}
	sort.Sort(list)

	/*var buf bytes.Buffer
	for _,v := range list{
		buf.WriteString(fmt.Sprintf("%s,%d\n", v.word, v.count))
	}
	err = ioutil.WriteFile(resultfile, []byte(buf.String()), 0644)
	if err != nil{
		fmt.Println(err)
	}*/

	return list
}

func getWordCount(text string, ch chan map[string]int){
	wordMap := make(map[string]int,0)
	wordIdx := 0
	word := false
	for i,v := range text{
		if (v>=65 && v<=90)||(v>=97 && v<= 122) {
			if !word {
				word = true
				wordIdx = i
			}
		}else{
				if word{
					wordMap[text[wordIdx:i]]++
					word = false
				}
		}
		if word{
			wordMap[text[wordIdx:]]++
		}
		ch <- wordMap
	}
}

func splitText(text string, n int) []string{
	parts := make([]string, n)
	length := len(text)
	prepos := 0
	for i:=0;i<n-1;i++{
		j := length / n *(i+1)
		for string(text[j]) != " "{
			j++
		}
		parts[i] = text[prepos:j]
		prepos = j
	}
	parts[n-1] = text[prepos:]
	return parts
}

func getChineseText(textfile string) WordCountList{
	data,_ := ioutil.ReadFile(textfile)
	length := len(data)
	text := string(data)
	wordMap := splitCnText(data, &text, length)
	list := make(WordCountList, 0)
	for k,v := range wordMap{
		wordcount := &WordCount{k,v}
		list = append(list, wordcount)
	}
	sort.Sort(list)
	return list
}

func splitCnText(data []byte, text *string, n int) map[string]int{
	wordMap := make(map[string]int, 0)
	var i, x, y int = 0,0,0
	eword := false
	estr := ""
	//cstr2, cstr3 := "",""
	for i < n-2{
		if (data[i] >=65 && data[i] <=90)||(data[i] >=97 && data[i] <= 122){
			if !eword{
				eword = true
				estr = ""
				estr += string(data[i])
			}
			i++
		}else if data[i] < 128{
			if eword{
				wordMap[estr]++
				eword = false
			}
			i++
		}else{ //if data[i] > 128
			x = i + 3
			if data[x] > 128{
				y = x + 3
				if data[y] > 128{
					cstr := string(data[i:i+3]) + string(data[x:x+3]) + string(data[y:y+3])
					wordMap[cstr]++
				}else{
					cstr := string(data[i:i+3]) + string(data[x:x+3])
					wordMap[cstr]++
				}
			}
			i += 3
			/*for x < n-2{
				if data[x] > 128{
					cstr2 = string(data[i:i+3]) + string(data[x:x+3])
					if j := strings.Count(*text, cstr2); j >= 1{
						wordMap[cstr2] = j
					}
					y = x+3
					for y < n-2{
						if data[y] > 128{
							cstr3 = cstr2 + string(data[y:y+3])
						}else{
							break
						}
					}
					break
				}else{
					x++
					cstr2 = ""
				}
			}
			if j := strings.Count(*text, cstr3); j >= 1{
				wordMap[cstr3] = j
			}
			i += 3*/
		}
	}
	if eword{
		estr += string(data[n-2:n])
		wordMap[estr]++
	}
	return wordMap
}
