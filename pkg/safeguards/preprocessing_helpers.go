package safeguards

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// Given a Helm chart directory or file, renders all templates and writes them to the specified directory
func RenderHelmChart(isFile bool, mainChartPath, tempDir string) ([]ManifestFile, error) {
	if isFile { // Get the directory that the Chart.yaml lives in
		mainChartPath = filepath.Dir(mainChartPath)
	}

	mainChart, err := loader.Load(mainChartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load main chart: %s", err)
	}

	loadedCharts := make(map[string]*chart.Chart) // map of chart path to chart object
	loadedCharts[mainChartPath] = mainChart

	// Load subcharts and dependencies
	for _, dep := range mainChart.Metadata.Dependencies {
		// Resolve the chart path based on the main chart's directory
		chartPath := filepath.Join(mainChartPath, dep.Repository[len("file://"):])
		chartPath = filepath.Clean(chartPath)

		subChart, err := loader.Load(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load chart: %s", err)
		}
		loadedCharts[chartPath] = subChart
	}

	var manifestFiles []ManifestFile
	for chartPath, chart := range loadedCharts {
		valuesPath := filepath.Join(chartPath, "values.yaml") // Enforce that values.yaml must be at same level as Chart.yaml
		mergedValues, err := getValues(chart, valuesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load values: %s", err)
		}
		e := engine.Engine{Strict: true}
		renderedFiles, err := e.Render(chart, mergedValues)
		if err != nil {
			return nil, fmt.Errorf("failed to render chart: %s", err)
		}

		// Write each rendered file to the output directory with the same name as in templates/
		for filePath, content := range renderedFiles {
			outputFilePath := filepath.Join(tempDir, filepath.Base(filePath))
			if err := os.WriteFile(outputFilePath, []byte(content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write manifest file: %s", err)
			}
			manifestFiles = append(manifestFiles, ManifestFile{Name: filepath.Base(filePath), Path: outputFilePath})
		}
	}

	return manifestFiles, nil
}

// Returns values from values.yaml and release options specified in values.yaml
func getValues(chart *chart.Chart, valuesPath string) (chartutil.Values, error) {
	// Load values file
	valuesFile, err := os.ReadFile(valuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %s", err)
	}

	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(valuesFile, &vals); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %s", err)
	}

	mergedValues, err := getReleaseOptions(chart, vals)
	return mergedValues, err
}

func getReleaseOptions(chart *chart.Chart, vals map[string]interface{}) (chartutil.Values, error) {
	// Extract release options from values
	releaseName, ok := vals["releaseName"].(string)
	if !ok || releaseName == "" {
		return nil, fmt.Errorf("releaseName not found or empty in values.yaml")
	}

	releaseNamespace, ok := vals["releaseNamespace"].(string)
	if !ok || releaseNamespace == "" {
		return nil, fmt.Errorf("releaseNamespace not found or empty in values.yaml")
	}

	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: releaseNamespace,
	}

	// Combine chart values with release options
	config := chartutil.Values(vals)
	mergedValues, err := chartutil.ToRenderValues(chart, config, options, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %s", err)
	}

	return mergedValues, nil
}
