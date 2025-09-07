package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	var lambdaName string
	if len(os.Args) > 1 {
		lambdaName = os.Args[1]
	}

	// Costruisci l'immagine Docker
	var buildCmd *exec.Cmd
	if lambdaName == "" {
		buildCmd = exec.Command("docker", "build", "--network=host", "-t", "golambda", ".", "-f", "deploy/docker/lambda.Dockerfile")
	} else {
		buildCmd = exec.Command("docker", "build", "--network=host", "-t", "golambda", ".", "-f", "deploy/docker/lambda.Dockerfile", "--build-arg", "LAMBDA_PATH="+lambdaName)
	}
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Println("Errore build Docker:", err)
		os.Exit(1)
	}

	// Crea container temporaneo
	exec.Command("docker", "rm", "-f", "tmpcontainer").Run() // Rimuovi se già esiste
	createCmd := exec.Command("docker", "create", "--name", "tmpcontainer", "golambda")
	if err := createCmd.Run(); err != nil {
		fmt.Println("Errore creazione container:", err)
		os.Exit(1)
	}

	// Crea cartella temporanea
	os.RemoveAll("lambda_tmp")
	os.Mkdir("lambda_tmp", 0755)

	// Copia i file dal container
	cpCmd := exec.Command("docker", "cp", "tmpcontainer:/lambda/.", "lambda_tmp")
	cpCmd.Stdout = os.Stdout
	cpCmd.Stderr = os.Stderr
	if err := cpCmd.Run(); err != nil {
		fmt.Println("Errore copia file dal container:", err)
		os.Exit(1)
	}

	// Cancella e ricrea la cartella di destinazione
	os.RemoveAll("lambda")
	os.Mkdir("lambda", 0755)

	// Copia solo i file .zip mantenendo la struttura
	filepath.Walk("lambda_tmp", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".zip" {
			rel, _ := filepath.Rel("lambda_tmp", path)
			dest := filepath.Join("lambda", rel)
			os.MkdirAll(filepath.Dir(dest), 0755)
			srcFile, _ := os.Open(path)
			defer srcFile.Close()
			dstFile, _ := os.Create(dest)
			defer dstFile.Close()
			_, err := dstFile.ReadFrom(srcFile)
			return err
		}
		return nil
	})

	// Pulisci la cartella temporanea
	os.RemoveAll("lambda_tmp")

	// Rimuovi il container temporaneo
	exec.Command("docker", "rm", "tmpcontainer").Run()

}
