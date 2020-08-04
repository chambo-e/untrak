package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource is a minimal description of a kubernetes object
type Resource struct {
	Kind              string `yaml:"kind,omitempty"`
	APIVersion        string `yaml:"apiVersion,omitempty"`
	metav1.ObjectMeta `yaml:"metadata,omitempty"`
	Items             []*Resource `yaml:"items,omitempty"`
}

type Resources []Resource

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func getPaths(root string) ([]string, error) {
	if info, err := os.Stat(root); err == nil && info.IsDir() {
		YMLPaths, err := WalkMatch(root, "*.yml")
		if err != nil {
			return []string{}, err
		}
		YAMLPaths, err := WalkMatch(root, "*.yaml")
		if err != nil {
			return []string{}, err
		}

		return append(YMLPaths, YAMLPaths...), nil
	}
	return []string{root}, nil
}

func loadLocalResources(root string) (*Resources, error) {
	paths, err := getPaths(root)
	if err != nil {
		return nil, err
	}

	resources := Resources{}

	for _, path := range paths {
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		r := bytes.NewReader(yamlFile)
		dec := yaml.NewDecoder(r)

		var t Resource
		for dec.Decode(&t) == nil {
			resources = append(resources, t)
		}
	}

	return &resources, nil
}

func main() {
	// Flags, command line parameters
	// var cfgPathOpt = flag.String("config", "./untrak.yaml", "untrak Config Path")
	// var outputOpt = flag.String("o", "text", "Output format")
	flag.Parse()
	args := flag.Args()

	root := "."
	if len(args) >= 1 {
		root = args[0]
	}

	localResources, err := loadLocalResources(root)
	if err != nil {
		log.Fatalln(err)
	}

	for _, res := range *localResources {
		fmt.Println(res.APIVersion, res.Kind, res.Name, res.Namespace)
	}

	remoteResources, err := loadRemoteResources()
	if err != nil {
		log.Fatalln(err)
	}

	for _, res := range *localResources {
		fmt.Println(res.APIVersion, res.Kind, res.Name, res.Namespace)
	}

	// fmt.Println(paths)
	// fmt.Println(resources)

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	resourcesOut, err = getKubernetesResources(cfg.Out)
	// 	if err != nil {
	// 		log.Printf("[ERR] Failed to get Kubernetes resources (out): %v\n", err)
	// 		os.Exit(1)
	// 	}
	// }()

	// wg.Wait()

	// untrackedResources := listUntrackedResources(resourcesIn, resourcesOut, cfg.Exclude)
	// switch {
	// case *outputOpt == "text":
	// 	outputs.Text(untrackedResources)
	// case *outputOpt == "yaml":
	// 	outputs.YAML(untrackedResources)
	// default:
	// 	outputs.Text(untrackedResources)
	// }
}

// func getKubernetesResources(cfgs []*config.CommandConfig) ([]*Resource, error) {
// 	const yamlSeparator = "---\n"
// 	var resources []*Resource

// 	var wg sync.WaitGroup
// 	var mutex = &sync.Mutex{}

// 	for _, cfg := range cfgs {
// 		wg.Add(1)
// 		go func(cmd string, args ...string) {
// 			defer wg.Done()

// 			fullArgs := []string{}
// 			for _, arg := range args {
// 				if matches, err := filepath.Glob(arg); err == nil {
// 					fullArgs = append(fullArgs, matches...)
// 				}

// 			}

// 			fmt.Println(fullArgs)

// 			c := exec.Command(cmd, fullArgs...)
// 			var outb, errb bytes.Buffer
// 			c.Stdout = &outb
// 			c.Stderr = &errb
// 			err := c.Run()
// 			if err != nil {
// 				log.Fatal(err, errb.String())
// 			}
// 			stdoutDec := yaml.NewDecoder(&outb)
// 			fmt.Println(stdoutDec)
// 			for {
// 				tempResource := &Resource{}
// 				err := stdoutDec.Decode(tempResource)
// 				if err != nil && err != io.EOF {
// 					log.Printf("[ERR] Failed to decode yaml stream: %s\n", err.Error())
// 					os.Exit(1)
// 				}
// 				if err == io.EOF {
// 					break
// 				}
// 				if tempResource.Kind == "List" {
// 					mutex.Lock()
// 					resources = append(resources, tempResource.Items...)
// 					mutex.Unlock()
// 					continue
// 				}
// 				// Resource can be empty if yaml file has return lines, separators or comments
// 				// for example:
// 				// # empty resource
// 				// ---
// 				// ---
// 				// YAML decoder consider these lines valid but resource will be uninitialized
// 				if !tempResource.Empty() {
// 					mutex.Lock()
// 					resources = append(resources, tempResource)
// 					mutex.Unlock()
// 				}
// 			}
// 		}(cfg.Cmd, cfg.Args...)
// 	}
// 	wg.Wait()
// 	return resources, nil
// }

// func listUntrackedResources(in []*Resource, out []*Resource, kindExclude []string) []*Resource {
// 	var untrackedResources []*Resource
// 	for _, resourceOut := range out {
// 		// Resource is in the exlude list, skip it
// 		if utils.StringInListCaseInsensitive(kindExclude, resourceOut.Kind) {
// 			continue
// 		}
// 		found := false
// 		for _, resourceIn := range in {
// 			// If resource has been found in both IN an OUT, there is nothing to do
// 			if resourceOut.ID() == resourceIn.ID() {
// 				found = true
// 				break
// 			}
// 		}
// 		// If resource OUT is not found in IN, it is untracked
// 		if !found {
// 			untrackedResources = append(untrackedResources, resourceOut)
// 		}
// 	}

// 	return untrackedResources
// }
