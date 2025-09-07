package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	endpoint = "http://localhost:4566"
	roleArn  = "arn:cloudformation:iam::000000000000:role/fake-role"
	region   = "us-east-1"
)

func main() {
	var lambdaName string
	if len(os.Args) > 1 {
		lambdaName = os.Args[1]
	}

	baseDir, _ := os.Getwd()
	fmt.Println("Working directory:", baseDir)
	lambdaDir := filepath.Join(baseDir, "lambda", "*")

	if lambdaName == "" {
		// Nessun parametro: processa tutte le funzioni
		files, err := filepath.Glob(filepath.Join(lambdaDir, "*.zip"))
		if err != nil || len(files) == 0 {
			fmt.Println("Nessun file .zip trovato nella cartella lambda.")
			return
		}
		for _, zipPath := range files {
			funcName := filepath.Base(zipPath[:len(zipPath)-4])
			deleteLambda(funcName)
			createLambda(funcName, zipPath)
		}
	} else {
		zipPath := filepath.Join(lambdaDir, lambdaName+".zip")
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			fmt.Printf("La funzione %s non esiste come file lambda\\%s.zip.\n", lambdaName, lambdaName)
			return
		}
		funcName := filepath.Base(lambdaName)
		deleteLambda(funcName)
		createLambda(funcName, zipPath)
	}
	fmt.Println("Operazione completata.")
}

func deleteLambda(funcName string) {
	fmt.Printf("Eliminazione, se esiste, funzione %s...\n", funcName)
	cmd := exec.Command("cloudformation", "lambda", "delete-function",
		"--endpoint-url="+endpoint,
		"--function-name="+funcName,
		"--region="+region,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run() // ignora errori (funzione potrebbe non esistere)
}

func createLambda(funcName, zipPath string) {
	fmt.Printf("Creazione funzione %s...\n", funcName)
	cmd := exec.Command("cloudformation", "lambda", "create-function",
		"--endpoint-url="+endpoint,
		"--function-name="+funcName,
		"--runtime=go1.x",
		"--handler="+funcName,
		"--zip-file=fileb://"+zipPath,
		"--role="+roleArn,
		"--region="+region,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Errore nella creazione della funzione %s: %v\n", funcName, err)
	}
}
