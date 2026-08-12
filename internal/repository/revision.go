package repository

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const gitTimeout = 10 * time.Second
var fullSHA=regexp.MustCompile(`^[0-9a-f]{40}$`)

func CleanRevision(root string)(string,error){
	sha,err:=gitOutput(root,"rev-parse","HEAD");if err!=nil||!fullSHA.MatchString(sha){return "",errors.New("repository is not a Git checkout with a full lowercase commit revision")}
	status,err:=gitOutput(root,"status","--porcelain");if err!=nil{return "",errors.New("repository status check failed")};if status!=""{return "",errors.New("repository worktree is dirty")};return sha,nil
}

// Revision returns a commit-bound identifier. A dirty worktree is made explicit.
func Revision(root string) string {
	sha, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil || sha == "" { return "unknown" }
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil { return sha + "+status-unknown" }
	if status != "" { return sha + "+dirty" }
	return sha
}
func gitOutput(root string,args ...string)(string,error){ctx,cancel:=context.WithTimeout(context.Background(),gitTimeout);defer cancel();commandArgs:=append([]string{"-C",root},args...);command:=exec.CommandContext(ctx,"git",commandArgs...);var output bytes.Buffer;command.Stdout=&output;command.Stderr=&bytes.Buffer{};if err:=command.Run();err!=nil{return "",err};return strings.TrimSpace(output.String()),nil}
