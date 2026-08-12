package approval

import(
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type Manifest struct{SchemaVersion string `json:"schema_version"`;Repository string `json:"repository"`;FinalRevision string `json:"final_revision"`;BindingDigest string `json:"binding_digest"`;PolicyDigest string `json:"policy_digest"`;RequirementsDigest string `json:"requirements_digest"`;ChangeSetDigest string `json:"change_set_digest"`;TestDigest string `json:"test_attestations_digest"`;ReviewDigest string `json:"review_attestations_digest"`;DecisionDigest string `json:"decision_attestations_digest"`}
func Load(path string)(Manifest,error){var m Manifest;err:=strict(path,&m);return m,err}
func Verify(manifestPath,signaturePath,keyPath string,m Manifest)error{body,err:=os.ReadFile(manifestPath);if err!=nil{return err};signatureText,err:=os.ReadFile(signaturePath);if err!=nil{return err};keyText,err:=os.ReadFile(keyPath);if err!=nil{return err};signature,err:=base64.StdEncoding.DecodeString(string(bytesTrim(signatureText)));if err!=nil{return err};key,err:=base64.StdEncoding.DecodeString(string(bytesTrim(keyText)));if err!=nil{return err};if len(key)!=ed25519.PublicKeySize||!ed25519.Verify(ed25519.PublicKey(key),body,signature){return errors.New("approval signature verification failed")};return nil}
func Digest(path string)(string,error){body,err:=os.ReadFile(path);if err!=nil{return "",err};sum:=sha256.Sum256(body);return hex.EncodeToString(sum[:]),nil}
func strict(path string,target any)error{f,err:=os.Open(path);if err!=nil{return err};defer f.Close();d:=json.NewDecoder(f);d.DisallowUnknownFields();if err:=d.Decode(target);err!=nil{return err};var x any;if err:=d.Decode(&x);!errors.Is(err,io.EOF){return errors.New("unexpected trailing JSON")};return nil}
func bytesTrim(value []byte)[]byte{start,end:=0,len(value);for start<end&&(value[start]==' '||value[start]=='\n'||value[start]=='\r'||value[start]=='\t'){start++};for end>start&&(value[end-1]==' '||value[end-1]=='\n'||value[end-1]=='\r'||value[end-1]=='\t'){end--};return value[start:end]}
