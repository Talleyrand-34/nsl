
/*
  Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)
 
  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Affero General Public License as published
  by the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.
 
  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  GNU Affero General Public License for more details.
 
  You should have received a copy of the GNU Affero General Public License
  along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package format

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

func WriteDiagram(
	d2Script string,
	outputPath string,
	outputDiagramFile string,
	outputImageFile string,
) {
	outFileD2 := filepath.Join(outputPath, outputDiagramFile)
	err := os.WriteFile(outFileD2, []byte(d2Script), 0600)
	if err != nil {
		panic("Failed to write D2lang script file: " + err.Error())
	}
	ruler, _ := textmeasure.NewRuler()
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}
	renderOpts := &d2svg.RenderOpts{
		Pad:     go2.Pointer(int64(5)),
		ThemeID: &d2themescatalog.GrapeSoda.ID,
	}
	ctx := log.WithDefault(context.Background())

	opts := &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: layoutResolver,
	}
	graph, _, err := d2lib.Compile(ctx, d2Script, opts, renderOpts)
	if err != nil {
		panic("Failed to compile D2 script: " + err.Error())
	}

	outFilesvg := filepath.Join(outputPath, outputImageFile)
	outSvg, err := d2svg.Render(graph, renderOpts)
	if err != nil {
		panic("Failed to render SVG: " + err.Error())
	}

	err = os.WriteFile(outFilesvg, outSvg, 0600)
	if err != nil {
		panic("Failed to write SVG file: " + err.Error())
	}

	fmt.Println("Diagram successfully generated in", outFilesvg)
}
