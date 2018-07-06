package tasks

import (
	"fmt"
	"net/http"
)

func LetsEncryptValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, "ZCLfT3oIOdBK0iUF28viK2IEvmjJ46_8NzBEE0F6jxA.oKzw5QYkN1q8zhhAe-jdS_VxeEiIqz4MC7pSvnuwGq4")
	return
}
