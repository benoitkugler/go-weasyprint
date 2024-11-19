package pdf

import (
	"strings"
	"testing"

	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/svg"
)

func drawStandaloneSVG(t *testing.T, input string, outFile string) {
	dst := newGroup(newCache(), 0, 0, 600, 600)
	dst.Transform(matrix.New(1, 0, 0, -1, 0, 600)) // SVG use "mathematical conventions"
	img, err := svg.Parse(strings.NewReader(input), "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	img.Draw(&dst, 600, 600, nil)

	var out model.Document
	var page model.PageObject

	dst.stream.ApplyToPageObject(&page, false)
	out.Catalog.Pages.Kids = append(out.Catalog.Pages.Kids, &page)

	if err := out.WriteFile(outFile, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSVG(t *testing.T) {
	input := `
		<?xml version="1.0" encoding="iso-8859-1"?>
	<!-- Generator: Adobe Illustrator 19.0.0, SVG Export Plug-In . SVG Version: 6.00 Build 0)  -->
	<svg version="1.1" id="Capa_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
		viewBox="0 0 512 512" style="enable-background:new 0 0 512 512;" xml:space="preserve">
	<g>
		<rect x="136" y="232" style="fill:#E3E6E9;" width="16" height="48"/>
		<rect x="360" y="232" style="fill:#E3E6E9;" width="16" height="48"/>
		<path style="fill:#E3E6E9;" d="M320,144h-24v16c0,1.105-0.056,2.196-0.165,3.272c-0.246,2.427-0.773,4.769-1.531,7.004
			c-0.141,0.415-0.289,0.826-0.446,1.234c-0.207,0.536-0.428,1.065-0.663,1.587c-0.319,0.71-0.656,1.41-1.024,2.091
			c-0.076,0.14-0.157,0.276-0.234,0.415c-0.798,1.425-1.702,2.781-2.701,4.061c-0.304,0.39-0.617,0.772-0.938,1.147
			c-0.536,0.625-1.089,1.235-1.671,1.817c-0.018,0.018-0.037,0.034-0.054,0.052c-0.652,0.649-1.333,1.267-2.039,1.858
			c-0.613,0.513-1.247,1.001-1.898,1.468c-0.773,0.555-1.57,1.077-2.393,1.563c-0.471,0.278-0.95,0.543-1.436,0.798
			c-0.724,0.379-1.461,0.735-2.216,1.058c-0.212,0.091-0.428,0.175-0.643,0.261c-3.21,1.293-6.674,2.089-10.302,2.273
			C265.101,191.986,264.552,192,264,192h-8h-8c-17.673,0-32-14.327-32-32v-16h-24c-30.875,0-56,25.125-56,56v16h16
			c0-8.837,7.164-16,16-16h8v16v16v48h16c0-4.422,3.578-8,8-8s8,3.578,8,8h16c0-4.422,3.578-8,8-8s8,3.578,8,8h16h80v-48v-16v-16h8
			c8.836,0,16,7.163,16,16h16v-16C376,169.125,350.875,144,320,144z M296,240h-16c-4.422,0-8-3.578-8-8s3.578-8,8-8h16
			c4.422,0,8,3.578,8,8S300.422,240,296,240z M296,216h-16c-4.422,0-8-3.578-8-8s3.578-8,8-8h16c4.422,0,8,3.578,8,8
			S300.422,216,296,216z"/>
		<path style="fill:#E3E6E9;" d="M360,331.542c-6.623,5.97-14.835,10.164-24,11.713V336v-24h-80h-16v8c0,4.422-3.578,8-8,8
			s-8-3.578-8-8v-8h-16v8c0,4.422-3.578,8-8,8s-8-3.578-8-8v-8h-16v24v7.255c-9.165-1.549-17.377-5.744-24-11.713V296h-16
			c0,24.047,17.773,44.016,40.875,47.469l14.19,113.545c2.874-0.656,5.862-1.014,8.935-1.014h40c4.418,0,8,3.582,8,8v-96
			c0-4.414,3.586-8,8-8s8,3.586,8,8v96c0-4.418,3.582-8,8-8h40c3.073,0,6.06,0.358,8.935,1.014l14.19-113.545
			C358.226,340.015,376,320.047,376,296h-16V331.542z"/>
		<path style="fill:#B6B8BE;" d="M344,304c0,0.276-0.014,0.549-0.041,0.818c-0.041,0.407-0.122,0.802-0.223,1.189
			c-0.05,0.193-0.104,0.383-0.168,0.569c-0.789,2.319-2.611,4.155-4.92,4.966c-0.83,0.292-1.719,0.458-2.648,0.458v24v7.255
			c9.165-1.549,17.377-5.744,24-11.713V296h-16V304z"/>
		<path style="fill:#B6B8BE;" d="M344,200h-8v16h24C360,207.163,352.836,200,344,200z"/>
		<rect x="336" y="232" style="fill:#B6B8BE;" width="24" height="48"/>
		<path style="fill:#B6B8BE;" d="M168,304v-8h-16v35.542c6.623,5.97,14.835,10.164,24,11.713V336v-24
			C171.582,312,168,308.418,168,304z"/>
		<rect x="152" y="232" style="fill:#B6B8BE;" width="24" height="48"/>
		<path style="fill:#B6B8BE;" d="M176,200h-8c-8.836,0-16,7.163-16,16h24V200z"/>
		<path style="fill:#FF9300;" d="M200,120h56V0c-44.109,0-80,35.891-80,80c0,15.898,4.656,31.266,13.477,44.445
			c0.682,1.019,1.583,1.841,2.609,2.443C192.628,122.998,195.96,120,200,120z"/>
		<path style="fill:#FFCF00;" d="M312,120c4.042,0,7.374,3,7.915,6.893c1.026-0.601,1.927-1.421,2.609-2.44
			C331.336,111.281,336,95.906,336,80c0-44.109-35.891-80-80-80v120H312z"/>
		<path style="fill:#888693;" d="M224,136h16h16v-8v-8h-56c-4.04,0-7.372,2.998-7.914,6.889C192.035,127.253,192,127.622,192,128v16
			h24C216,139.582,219.582,136,224,136z"/>
		<path style="fill:#FF4F19;" d="M296,200h-16c-4.422,0-8,3.578-8,8s3.578,8,8,8h16c4.422,0,8-3.578,8-8S300.422,200,296,200z"/>
		<path style="fill:#FF4F19;" d="M296,224h-16c-4.422,0-8,3.578-8,8s3.578,8,8,8h16c4.422,0,8-3.578,8-8S300.422,224,296,224z"/>
		<path style="fill:#B6B8BE;" d="M312,120h-56v8v8h16h16c4.418,0,8,3.582,8,8h24v-16c0-0.377-0.035-0.744-0.085-1.107
			C319.374,123,316.042,120,312,120z"/>
		<path style="fill:#5C546A;" d="M240,152v-16h-16c-4.418,0-8,3.582-8,8v16c0,17.673,14.327,32,32,32h8v-24
			C247.163,168,240,160.836,240,152z"/>
		<path style="fill:#5C546A;" d="M296,160c0,1.105-0.056,2.196-0.165,3.272C295.944,162.196,296,161.104,296,160z"/>
		<path style="fill:#5C546A;" d="M276.592,189.424c-0.212,0.091-0.428,0.175-0.643,0.261
			C276.163,189.599,276.379,189.515,276.592,189.424z"/>
		<path style="fill:#5C546A;" d="M284.534,184.537c-0.613,0.513-1.247,1.001-1.898,1.468
			C283.287,185.538,283.921,185.051,284.534,184.537z"/>
		<path style="fill:#5C546A;" d="M295.835,163.272c-0.246,2.427-0.773,4.769-1.531,7.004
			C295.062,168.041,295.588,165.698,295.835,163.272z"/>
		<path style="fill:#5C546A;" d="M280.244,187.568c-0.471,0.278-0.95,0.543-1.436,0.798
			C279.294,188.111,279.773,187.846,280.244,187.568z"/>
		<path style="fill:#5C546A;" d="M289.236,179.664c-0.304,0.39-0.617,0.772-0.938,1.147
			C288.619,180.436,288.932,180.053,289.236,179.664z"/>
		<path style="fill:#5C546A;" d="M293.858,171.51c-0.207,0.536-0.428,1.065-0.663,1.587
			C293.43,172.575,293.651,172.046,293.858,171.51z"/>
		<path style="fill:#5C546A;" d="M292.171,175.188c-0.076,0.14-0.157,0.276-0.234,0.415
			C292.014,175.464,292.096,175.327,292.171,175.188z"/>
		<path style="fill:#5C546A;" d="M286.627,182.627c-0.018,0.018-0.037,0.034-0.054,0.052
			C286.591,182.661,286.61,182.645,286.627,182.627z"/>
		<path style="fill:#888693;" d="M288,136h-16v16c0,8.837-7.163,16-16,16v24h8c0.552,0,1.101-0.014,1.647-0.042
			c3.628-0.184,7.092-0.98,10.302-2.273c0.215-0.086,0.431-0.17,0.643-0.261c0.755-0.324,1.493-0.68,2.216-1.058
			c0.486-0.254,0.965-0.519,1.436-0.798c0.822-0.486,1.62-1.008,2.393-1.563c0.65-0.467,1.285-0.954,1.898-1.468
			c0.705-0.591,1.387-1.209,2.039-1.858c0.018-0.018,0.037-0.034,0.054-0.052c0.582-0.582,1.135-1.192,1.671-1.817
			c0.321-0.375,0.634-0.757,0.938-1.147c0.999-1.28,1.903-2.636,2.701-4.061c0.078-0.139,0.159-0.275,0.234-0.415
			c0.368-0.681,0.705-1.381,1.024-2.091c0.234-0.522,0.456-1.051,0.663-1.587c0.157-0.407,0.305-0.819,0.446-1.234
			c0.758-2.235,1.284-4.578,1.531-7.004c0.109-1.076,0.165-2.167,0.165-3.272v-16C296,139.581,292.418,136,288,136z"/>
		<path style="fill:#FF4F19;" d="M256,168c8.837,0,16-7.163,16-16v-16h-16v8V168z"/>
		<path style="fill:#E5001E;" d="M240,152c0,8.837,7.163,16,16,16v-24v-8h-16V152z"/>
		<rect x="336" y="216" style="fill:#E5001E;" width="24" height="16"/>
		<rect x="152" y="216" style="fill:#E5001E;" width="24" height="16"/>
		<rect x="360" y="216" style="fill:#FF4F19;" width="16" height="16"/>
		<rect x="136" y="216" style="fill:#FF4F19;" width="16" height="16"/>
		<path style="fill:#5C546A;" d="M152,296h16v8c0,4.418,3.582,8,8,8h16v-32h-16h-24V296z"/>
		<rect x="136" y="280" style="fill:#888693;" width="16" height="16"/>
		<path style="fill:#5C546A;" d="M344,288v8h16v-16h-24C340.418,280,344,283.581,344,288z"/>
		<rect x="360" y="280" style="fill:#888693;" width="16" height="16"/>
		<rect x="208" y="280" style="fill:#5C546A;" width="16" height="32"/>
		<rect x="240" y="280" style="fill:#5C546A;" width="16" height="32"/>
		<path style="fill:#5C546A;" d="M343.568,306.576c-0.789,2.319-2.611,4.155-4.92,4.966
			C340.957,310.731,342.779,308.895,343.568,306.576z"/>
		<path style="fill:#5C546A;" d="M343.736,306.006c0.1-0.387,0.181-0.781,0.223-1.189
			C343.917,305.225,343.836,305.619,343.736,306.006z"/>
		<path style="fill:#888693;" d="M336,312c0.93,0,1.818-0.167,2.648-0.458c2.309-0.811,4.131-2.647,4.92-4.966
			c0.064-0.187,0.118-0.377,0.168-0.569c0.1-0.387,0.181-0.781,0.223-1.189c0.027-0.269,0.041-0.542,0.041-0.818v-8v-8
			c0-4.418-3.582-8-8-8h-80v32H336z"/>
		<path style="fill:#FF9300;" d="M192,320c0,4.422,3.578,8,8,8s8-3.578,8-8v-8v-32c0-4.422-3.578-8-8-8s-8,3.578-8,8v32V320z"/>
		<path style="fill:#E5001E;" d="M224,320c0,4.422,3.578,8,8,8s8-3.578,8-8v-8v-32c0-4.422-3.578-8-8-8s-8,3.578-8,8v32V320z"/>
		<path style="fill:#888693;" d="M160.432,506.576c0.789,2.319,2.611,4.155,4.92,4.966
			C163.043,510.731,161.221,508.895,160.432,506.576z"/>
		<path style="fill:#888693;" d="M247.959,504.818c-0.041,0.407-0.122,0.802-0.223,1.189
			C247.836,505.619,247.917,505.225,247.959,504.818z"/>
		<path style="fill:#888693;" d="M160.264,506.006c-0.1-0.387-0.181-0.781-0.223-1.189
			C160.083,505.225,160.164,505.619,160.264,506.006z"/>
		<path style="fill:#888693;" d="M242.648,511.542c2.309-0.811,4.131-2.647,4.92-4.966
			C246.779,508.895,244.957,510.731,242.648,511.542z"/>
		<path style="fill:#888693;" d="M240,456h-40c-3.073,0-6.06,0.358-8.935,1.014C173.277,461.074,160,476.981,160,496
			c0-2.74,0.281-5.415,0.806-8H248v-24C248,459.581,244.418,456,240,456z"/>
		<path style="fill:#888693;" d="M264.264,506.006c-0.1-0.387-0.181-0.781-0.223-1.189
			C264.083,505.225,264.164,505.619,264.264,506.006z"/>
		<path style="fill:#888693;" d="M346.648,511.542c2.309-0.811,4.131-2.647,4.92-4.966
			C350.779,508.895,348.957,510.731,346.648,511.542z"/>
		<path style="fill:#888693;" d="M351.958,504.818c-0.041,0.407-0.122,0.802-0.223,1.189
			C351.836,505.619,351.917,505.225,351.958,504.818z"/>
		<path style="fill:#888693;" d="M264.432,506.576c0.789,2.319,2.611,4.155,4.92,4.966
			C267.043,510.731,265.221,508.895,264.432,506.576z"/>
		<path style="fill:#888693;" d="M272,456c-4.418,0-8,3.582-8,8v24h87.194c0.525,2.585,0.806,5.26,0.806,8
			c0-19.018-13.277-34.925-31.065-38.986C318.06,456.358,315.073,456,312,456H272z"/>
		<path style="fill:#B6B8BE;" d="M160,496v8c0,0.276,0.014,0.549,0.041,0.818c0.041,0.407,0.122,0.802,0.223,1.189
			c0.05,0.193,0.104,0.383,0.168,0.569c0.789,2.319,2.611,4.155,4.92,4.966c0.83,0.292,1.719,0.458,2.648,0.458h72
			c0.93,0,1.818-0.167,2.648-0.458c2.309-0.811,4.131-2.647,4.92-4.966c0.064-0.187,0.118-0.377,0.168-0.569
			c0.1-0.387,0.181-0.781,0.223-1.189c0.027-0.269,0.041-0.542,0.041-0.818v-16h-87.194C160.281,490.585,160,493.26,160,496z"/>
		<path style="fill:#B6B8BE;" d="M264,488v16c0,0.276,0.014,0.549,0.041,0.818c0.041,0.407,0.122,0.802,0.223,1.189
			c0.05,0.193,0.104,0.383,0.168,0.569c0.789,2.319,2.611,4.155,4.92,4.966c0.83,0.292,1.719,0.458,2.648,0.458h72
			c0.93,0,1.818-0.167,2.648-0.458c2.309-0.811,4.131-2.647,4.92-4.966c0.064-0.187,0.118-0.377,0.168-0.569
			c0.1-0.387,0.181-0.781,0.223-1.189c0.027-0.269,0.041-0.542,0.041-0.818v-8c0-2.74-0.281-5.415-0.806-8H264z"/>
	</g >
	</svg>
	`
	drawStandaloneSVG(t, input, "/tmp/svg_test.pdf")
}

func TestSVGVGradient(t *testing.T) {
	input := `
	<svg width="300px" height="300px" xmlns="http://www.w3.org/2000/svg">
	        <defs>
	          <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
	            gradientUnits="objectBoundingBox">
	            <stop stop-color="blue" offset="5%"></stop>
	            <stop stop-color="red" offset="30%"></stop>
	            <stop stop-color="green" offset="50%"></stop>
	          </linearGradient>
	        </defs>
	        <rect x="0" y="0" width="50" height="50" fill="url(#grad)" />
	      </svg>
	`
	drawStandaloneSVG(t, input, "/tmp/svg_gradient_test.pdf")
}

func TestSVGMask(t *testing.T) {
	input := `
	<svg viewBox="-10 -10 150 150">
	<mask id="myMask">
		<!-- Everything under a white pixel will be visible -->
		<rect x="0" y="0" width="100" height="100" fill="white" />

		<!-- Everything under a black pixel will be invisible -->
		<path d="M10,35 A20,20,0,0,1,50,35 A20,20,0,0,1,90,35 Q90,65,50,95 Q10,65,10,35 Z" fill="black" />
	</mask>

	<polygon points="-10,110 110,110 110,-10" fill="orange" />

	<!-- with this mask applied, we "punch" a heart shape hole into the circle -->
	<circle cx="50" cy="50" r="50" mask="url(#myMask)" />
	</svg>
	`
	drawStandaloneSVG(t, input, "/tmp/mask.pdf")
}

func TestSVGGradient(t *testing.T) {
	input := `
	<svg width="500" height="500"
		xmlns="http://www.w3.org/2000/svg"
		xmlns:xlink="http://www.w3.org/1999/xlink">


	<radialGradient id="rg2" cx="50%" cy="50%"   r="40%" gradientUnits="objectBoundingBox"  
	spreadMethod="repeat" >
	<stop offset="10%" stop-color="goldenrod" />
	<stop offset="30%" stop-color="seagreen" />
	<stop offset="50%" stop-color="cyan" />
	<stop offset="70%" stop-color="black" />
	<stop offset="100%" stop-color="orange" />
	</radialGradient>

	 


	<ellipse cx="300" cy="150" rx="120" ry="100"  style="fill:url(#rg2)" /> 

	</svg>
	`
	drawStandaloneSVG(t, input, "/tmp/gradient.pdf")
}
