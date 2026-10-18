package main
import "testing"
func TestAllowed(t *testing.T) {
 if !allowed([]string{"ec2","s3"},"s3") { t.Fatal("expected match") }
 if allowed([]string{"ec2","s3"},"rds") { t.Fatal("unexpected match") }
}
