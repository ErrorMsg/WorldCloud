package main

import "github.com/fogleman/gg"

func main(){
	const S = 1024
	dc := gg.NewContext(S,S)
	//dc.SetRGBA(0,0,0,0.1)
	for i:=0;i<360;i+=15{
		dc.SetRGBA(float64(255-i),float64(i),float64(255-i),0.1)
		dc.Push()
		dc.RotateAbout(gg.Radians(float64(i)),S/2,S/2)
		dc.DrawEllipse(S/2,S/2,S*7/16,S/8)
		dc.Fill()
		dc.Pop()
	}
	dc.SavePNG("out.png")
}

//////////
func generateCloud(){
	textList := []string{"aaa","bbb","ccc","ddd"}
	angles := []int{0, 15, -15, 90}
	colors := []*color.RGBA{
		&color.RGBA{0x0,0x60,0x30,0xff},
		&color.RGBA{0x60,0x0,0x0,0xff},
	}
	cloud := NewCloud(60, 8, "xxx.ttf", "xxx.png", textList, angles, colors, "out.png")
	cloud.Generate()
}

func main(){
	generateCloud()
}

type Cloud struct{
	MaxFontSize float64
	MinFontSize float64
	FontPath string
	OutlineImgPath string
	MeasureDc *gg.Context
	DrawDc *gg.Context
	TextList []string
	Angles []int
	Colors []*color.RGBA
	OutImgPath string
	worldMap *WorldMap
}

func NewCloud(maxFontSize, minFontSize float64, fontPath string, imgPath string, 
	textList ]\string, angles []int, colors []*color.RGBA, outImgPath string) *Cloud{
	cloud := &Cloud{
		MaxFontSize: maxFontSize,
		MinFontSize: minFontSize,
		FontPath: fontPath,
		OutlineImgPath: imgPath,
		TextList: textList,
		Angles: angles,
		Colors: colors,
		OutImgPath: outImgPath,
	}
	worldMap := TwoByBitmap(imgPath)
	cloud.worldMap = worldMap
	drawDc := gg.NewContext(worldMap.RealImageWidth, worldMap.RealImageHeight)
	drawDc.SetRGB(1,1,1)
	drawDc.Clear()
	drawDc.SetRGB(0,0,0)
	cloud.DrawDc = drawDc
	if err := drawDc.LoadFontFace(fontPath, cloud.MaxFontSize); err != nil{
		panic(err)
	}
	cloud.ResetMeasureDc(cloud.MaxFontSize)
	return cloud
}


func (this *Cloud) Generate(){
	curFontSize := this.MaxFontSize
	curTextIdx := 0x0
	colorIdx := 0x0
	checkRet := &CheckResult{}
	itemGrid := &Grid{}
	bigestSizeCnt := 0
	for {
		var msg string = this.TextList[curTextIdx]
		curTextIdx++
		curTextIdx = curTextIdx % len(this.TextList)
		color := this.Colors[colorIdx]
		colorIdx++
		colorIdx = colorIdx % len(this.Colors)
		this.DrawDc.SetRGB(float64(color.R), float64(color.G), float64(color.B))
		w, h, xscale, yscale := GetTextBound(this.MeasureDc, msg)
		itemGrid.XScale = int(xscale)
		itemGrid.YScale = int(yscale)
		if int(w)%2 != 0{
			w += XUNIT
		}
		if int(h)%2 != 0{
			h += YUNIT
		}
		positions, w1, h1 := TwoByBlock(int(w), int(h))
		itemGrid.Width = int(w1)
		itemGrid.Height = int(h1)
		itemGrid.positions = positions
		isFound := this.collisionCheck(0, this.worldMap, itemGrid, checkRet, this.Angles)
		if isFound{
			DrawText(this.DrawDc, msg, float64(checkRet.Xpos + itemGrid.XScale/2),
				float64(checkRet.Ypos + itemGrid.YScale/2), Angle2Pi(float64(checkRet.Angle)))
			if curFontSize == this.MaxFontSize{
				bigestSizeCnt++
				if bigestSizeCnt > len(this.TextList){
					this.UpdateFontSize(40)
				}
			}
		}else{
			curFontSize -= 3
			if curFontSize < this.MinFontSize{
				break
			}
			this.UpdateFontSize(curFontSize)
		}
	}
	this.DrawDc.SavePNG(this.OutImgPath)
}

func (this *Cloud) UpdateFontSize(curFontSize float64){
	this.DrawDc.SetFontSize(curFontSize)
	this.MeasureDc.SetFontSize(curFontSize)
}

func (this *Cloud) ResetMeasureDc(fontSize float64){
	measureDc := gg.NewContext(this.worldMap.RealImageWidth, this.worldMap.RealImageHeight)
	measureDc.SetRGBA(0,0,0,0)
	measureDc.Clear()
	this.MeasureDc = MeasureDc
	if err := measureDc.LoadFontFace(this.FontPath, fontSize); err != nil{
		panic(err)
	}
}

func (this *Cloud) collisionCheck(lastcheckAngle float64, worldMap *WorldMap, itemGrid *Grid, ret *CheckResult, tryAngles []int) bool{
	centerX := worldMap.Width/2
	centerY := worldMap.Height/2
	isFound := true
	xDistanceToCenter := 0
	yDistanceToCenter := 0
	tempXpos := 0
	tempYpos := 0
	angleMark := 0
	curAngleIdx := 0
	for angle := lastcheckAngle; angle <= DEGREE_360; angle++{
		curAngleIdx = 0
		angleMark = tryAngles[curAngleIdx]
		curAngleIdx++
		Rotate(itemGrid, float64(angleMark), centerX, centerY)
		xDiff := CosT(angle)*1
		yDiff := SinT(angle)*1
		tempXpos = 0
		tempYpos = 0
		xLeiji := xDiff
		yLeiji := yDiff
		xDistanceToCenter = 0
		yDistanceToCenter = 0
		result := IS_NOT_FIT
		for {
			result = IS_NOT_FIT
			if xDistanceToCenter != tempXpos || yDistanceToCenter != tempYpos{
				tempXpos = xDistanceToCenter
				tempYpos = yDistanceToCenter
				result = itemGrid.isFit(xDistanceToCenter, yDistanceToCenter, worldMap.Width, worldMap.Height, worldMap.collisionMap)
				if result == OUT_INDEX{
					if curAngleIdx < len(tryAngles){
						angleMark = tryAngles[curAngleIdx]
						curAngleIdx++
						Rotate(itemGrid, float64(angleMark), centerX, centerY)
						xLeiji = xDiff
						yLeiji = yDiff
						tempXpos = 0
						tempYpos = 0
						xDistanceToCenter = 0
						yDistanceToCenter = 0
					}else{
						ret.Angle = 0
						isFound = false
						break
					}
				}else if result == IS_FIT{
					isFound = true
					itemGrid.Fill(worldMap.Width, worldMap.Height, worldMap.collisionMap)
					ret.Angle = angleMark
					ret.Xpos = (xDistanceToCenter + centerX) * XUNIT
					ret.Ypos = (yDistanceToCenter + centerY) * XUNIT
					ret.LastCheckAngle = int(angle)
					break
				}
			}
			xLeiji += xDiff
			yLeiji += yDiff
			xDistanceToCenter = int(CeilT(xLeiji))
			yDistanceToCenter = int(CeilT(yLeiji))
		}
		if angle >= DEGREE_360{
			ret.Angle = 0
			isFound = false
			break
		}
		if result == IS_FIT{
			break
		}
	}
	return isFound
}

type CheckResult struct{
	Angle int
	Xpos int
	Ypos int
	LastCheckAngle int
}

const (
	IS_NOT_FIT = 1
	IS_FIT = 2
	OUT_INDEX = 3
	DEGREE_360 = 360
	DEGREE_180 = 180
	IS_EMPTY = 0
	XUNIT = 2
	YUNIT = 2
)

type Position struct{
	Xpos int
	Ypos int
	Value int
	XLeiji int
	YLeiji int
}

func NewPosition(xpos, ypos, value, xleiji, yleiji int) *Position{
	pos := &Position{
		Xpos: xpos,
		Ypos: ypos,
		Value: value,
		XLeiji: xleiji,
		YLeiji: yleiji,
	}
	return pos
}

type Grid struct{
	Width int
	Height int
	positions []*Position
	XScale int
	YScale int
}

func (this *Grid) isFit(xIncrement, yIncrement, width, height int, gridIntArray []int) int{
	for i:=0;i<this.Height;i++{
		for j:=0;j<this.Width;j++{
			index := i*this.Width + j
			pos := this.positions[index]
			if pos.Value != IS_EMPTY{
				pos.Xpos = pos.XLeiji + xIncrement
				pos.Ypos = pos.YLeiji + yIncrement
				if pos.Xpos < 0 || pos.Xpos >= width || pos.Ypos < 0 || pos.Ypos >= height{
					return OUT_INDEX
				}
				index = pos.Ypos*width + pos.Xpos
				if pos.Value != 0 && gridIntArray[index] == pos.Value{
					return IS_NOT_FIT
				}
			}
		}
	}
	return IS_FIT
}

func (this *Grid) setCollisionMap(collisionMap []int, width, height int){
	this.Width = width
	this.Height = height
	index := 0
	for y:=0;y<height;y++{
		for x:=0;x<width;x++{
			value := collisionMap[index]
			pos := NewPosition(x, y, value, 0, 0)
			this.positions = append(this.positions, pos)
			index++
		}
	}
}

func (this *Grid) Fill(gridIntArrayWidth, gridIntArrayHeight int, gridIntArray []int){
	for y:=0;y<this.Height;y++{
		for x:=0;x<this.Width;x++{
			index := y*this.Width + x
			pos := this.positions[index]
			index = pos.Ypos*gridIntArrayWidth + pos.Xpos
			if pos.Value != IS_EMPTY{
				gridIntArray[index] = pos.Value
			}
		}
	}
}

type WorldMap struct{
	Width int
	Height int
	CollisionMap []int
	RealImageWidth int
	RealImageHeight int
}

func (this *WorldMap) printMap(){
	for y:=0;y<this.Height;y++{
		str := ""
		for x:=0;x<this.Width;x++{
			idx := y*this.Width + x
			str = str + strconv.Itoa(this.CollisionMap[idx])
		}
		fmt.Println(str)
	}
}

func TwoByBitmap(imgpath string) *WorldMap{
	worldMap := &WorldMap{CollisionMap: make([]int, 0),}
	file,err := os.Open(imgpath)
	if err != nil{
		fmt.Println(err)
	}
	img,err := png.Decode(file)
	file.Close()
	bounds := img.Bounds()
	w := bounds.Size().X
	h := bounds.Size().Y
	worldMap.Width = w / XUNIT
	worldMap.Height = h / YUNIT
	worldMap.RealImageWidth = w
	worldMap.RealImageHeight = h
	for y:=0;y<worldMap.Height;y++{
		for x:=0;x<worldMap.Width;x++{
			color := img.At(x*XUNIT, y*YUNIT)
			_,_,_,alpha := color.RGBA()
			if alpha = 0{
				worldMap.CollisionMap = append(worldMap.CollisionMap, 1)
			}else{
				worldMap.CollisionMap = append(worldMap.CollisionMap, 0)
			}
		}
	}
	return worldMap
}

func TwoByBlock(width, height int) ([]*Position, int, int){
	positions := make([]*Position, 0)
	maxX := width / XUNIT
	maxY := height / YUNIT
	for y:=0;y<maxY;y++{
		for x:=0;x<maxX;x++{
			pos := NewPosition(x, y, IS_NOT_FIT, 0, 0)
			positions = append(positions, pos)
		}
	}
	return positions, maxX, maxY
}

func DrawText(dc *ggContext, text string, xpos, ypos, rotation float64){
	if rotation != 0{
		dc.RotateAbout(rotation, xpos, ypos)
	}
	dc.DrawStringAnchored(text, xpos, ypos, 0.5, 0.5)
	if rotation != 0{
		dc.RotateAbout(-rotation, xpos, ypos)
	}
}

func GetTextBound(measureDc *ggContext, text string) (w, h, xdiff, ydiff float64){
	measureDc.SetRGBA(0,0,0,0)
	measureDc.Clear()
	measureDc.SetRGB(0,0,0)
	measureDc.DrawStringAnchored(text, 375, 375, 0.5, 0.5)
	img := measureDc.Image()
	width := measureDc.Width()
	height := measureDc.Height()
	maxX := 0
	maxY := 0
	minX := 9999999
	minY := 9999999
	for y:=0;y<height;y++{
		for x:=0;x<width;x++{
			color := img.At(x, y)
			_,_,_,alpha := color.RGBA()
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
	w1, h1 := measureDc.MeasureString(text)
	wdiff := float64(maxX - minX)
	hdiff := float64(maxY - minY)
	xdiff = float64(w1 - wdiff)
	ydiff = float64(h1 - hdiff)
	return wdiff, hdiff, xdiff, ydiff
}

func Clear(dc *gg.Context){
	dc.SetRGBA(0,0,0,0)
	dc.Clear()
}

func Rotate(grid *Grid, angle float64, centerX, centerY int){
	maxX := grid.Width
	maxY := grid.Height
	width := maxX * XUNIT
	height := maxY * YUNIT
	halfX := width / 2 
	halfY := height / 2
	tempX := 0
	tempY := 0
	gridData := grid.positions
	sinPi := SinT(angle)
	cosPi := CosT(angle)
	for y:=0;y<maxY;y++{
		for x:=0;x<maxX;x++{
			index := y*maxX + x
			pos := gridData[index]
			pos.Xpos = x
			pos.Ypos = y
			pos.Xpos = pos.Xpos*XUNIT - halfX
			pos.Ypos = pos.Ypos*YUNIT - halfY
			tempX = pos.Xpos
			tempY = pos.Ypos
			pos.Xpos = (int)(float64(tempX)*cosPi - float64(tempY)*sinPi)
			pos.Ypos = (int)(float64(tempX)*sinPi - float64(tempY)*cosPi)
			pos.Xpos /= XUNIT
			pos.Ypos /= YUNIT
			pos.Xpos += centerX
			pos.Xpos += centerY
			pos.XLeiji = pos.Xpos
			pos.YLeiji = pos.Ypos
		}
	}
}

func CeilT(value float64) float64{
	return math.Ceil(value)
}

func CosT(angle float64) float64{
	angle = angle / DEGREE_180 * math.Pi
	return math.Cos(angle)
}

func SinT(angle float64) float64{
	angle = angle / DEGREE_180 * math.Pi
	return math.Sin(angle)
}

func Angle2Pi(angle float64) float64{
	return angle / DEGREE_180 * math.Pi
}

