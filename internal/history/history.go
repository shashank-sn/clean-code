package history
import("crypto/sha256";"encoding/hex";"encoding/json";"errors";"io";"os";"sort")
type Signal struct{Name string `json:"name"`;Value float64 `json:"value"`;Scale string `json:"scale"`;Provenance string `json:"provenance"`}
type Content struct{Repository string `json:"repository"`;Revision string `json:"revision"`;CreatedAt string `json:"created_at"`;Signals []Signal `json:"signals"`}
type Receipt struct{SchemaVersion string `json:"schema_version"`;Digest string `json:"digest"`;Content Content `json:"content"`}
type Report struct{SchemaVersion string `json:"schema_version"`;Receipts []Receipt `json:"receipts"`}
func Load(path string)(Receipt,error){f,err:=os.Open(path);if err!=nil{return Receipt{},err};defer f.Close();d:=json.NewDecoder(f);d.DisallowUnknownFields();var r Receipt;if err:=d.Decode(&r);err!=nil{return r,err};var x any;if err:=d.Decode(&x);!errors.Is(err,io.EOF){return r,errors.New("unexpected trailing JSON")};return r,nil}
func Digest(c Content)string{body,_:=json.Marshal(c);sum:=sha256.Sum256(body);return hex.EncodeToString(sum[:])}
func Build(receipts []Receipt)(Report,error){seen:=map[string]bool{};for _,r:=range receipts{c:=r.Content;if r.SchemaVersion!="1.0.0"||r.Digest!=Digest(c)||c.Repository==""||c.Revision==""||c.CreatedAt==""||seen[r.Digest]{return Report{},errors.New("receipt is invalid, duplicated, or tampered")};seen[r.Digest]=true;for _,s:=range c.Signals{if s.Name==""||s.Scale==""||s.Provenance==""{return Report{},errors.New("signal metadata is incomplete")}}};sort.Slice(receipts,func(i,j int)bool{a,b:=receipts[i].Content,receipts[j].Content;if a.CreatedAt==b.CreatedAt{return a.Revision<b.Revision};return a.CreatedAt<b.CreatedAt});return Report{SchemaVersion:"1.0.0",Receipts:receipts},nil}
