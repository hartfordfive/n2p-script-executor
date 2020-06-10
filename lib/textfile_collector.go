package lib

//const tagName = "textfile_collector"

// TextfileCollectorMetric is a struct which will be serialized to be written to a file
// to be picked up by the node_exporter
// type TextfileCollectorMetric struct {
// 	Name   string
// 	Labels map[string]string
// 	Value  float64
// }

// func (tc TextfileCollectorMetric) String() string {
// 	if len(tc.Labels) >= 1 {
// 		flattendLabels := []string{}
// 		for k, v := range tc.Labels {
// 			flattendLabels = append(flattendLabels, fmt.Sprintf("%s=\"%s\"", k, v))
// 		}

// 		if ValueCanBeInt(tc.Value) {
// 			return fmt.Sprintf("n2p_script_exec_%s{%s} %s", tc.Name, strings.Join(flattendLabels, ", "), convertToIntString(tc.Value))
// 		}
// 		return fmt.Sprintf("n2p_script_exec_%s{%s} %f", tc.Name, strings.Join(flattendLabels, ", "), tc.Value)
// 	}

// 	if ValueCanBeInt(tc.Value) {
// 		return fmt.Sprintf("n2p_script_exec_%s %s", tc.Name, convertToIntString(tc.Value))
// 	}
// 	return fmt.Sprintf("n2p_script_exec_%s %f", tc.Name, tc.Value)
// }
