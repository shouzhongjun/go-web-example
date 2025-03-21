package tools

import (
	"encoding/base64"
	"fmt"
	"testing"

	"goWebExample/api/protobuf/users/pb"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestGenerateLoginRequestData(t *testing.T) {
	// 创建登录请求
	req := &pb.LoginRequest{
		Username: "testuser",
		Password: "123456",
	}

	// 序列化为 protobuf
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	// 转换为 base64
	base64Data := base64.StdEncoding.EncodeToString(data)
	fmt.Printf("Base64 encoded request data:\n%s\n", base64Data)

	// 解码并验证
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}

	var decodedReq pb.LoginRequest
	if err := proto.Unmarshal(decodedData, &decodedReq); err != nil {
		t.Fatal(err)
	}

	if decodedReq.Username != req.Username || decodedReq.Password != req.Password {
		t.Errorf("Decoded data does not match original: got %v, want %v", decodedReq, req)
	}
}

func TestDecodeLoginResponse(t *testing.T) {
	// 您提供的 base64 编码响应数据
	base64Response := "CsQCZXlKaGJHY2lPaUpJVXpJMU5pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SnBjM01pT2lKbmIxZGxZa1Y0WVcxd2JHVWlMQ0psZUhBaU9qRTNOREkyTXpBek1qWXNJbTVpWmlJNk1UYzBNalUwTXpreU5pd2lhV0YwSWpveE56UXlOVFF6T1RJMkxDSjFjMlZ5WDJsa0lqb2lOVFV3WlRnME1EQXRaVEk1WWkwME1XUTBMV0UzTVRZdE5EUTJOalUxTkRRd01EQXdJaXdpZFhObGNtNWhiV1VpT2lKMFpYTjBkWE5sY2lJc0ltNXBZMnR1WVcxbElqb2libWxqYUdWdVp5SXNJbWx6WDJGa2JXbHVJanBtWVd4elpYMC41VFNTLXloZlMzLWJ6azNEUkNPMlkzNkJpT3ZhZTc3Vl9MWWNLY3NlTWFnEgduaWNoZW5nGhR0ZXN0dXNlckBleGFtcGxlLmNvbQ=="

	// 解码 base64
	protoData, err := base64.StdEncoding.DecodeString(base64Response)
	if err != nil {
		t.Fatalf("解码 base64 失败: %v", err)
	}

	// 解析 protobuf
	var resp pb.LoginResponse
	if err := proto.Unmarshal(protoData, &resp); err != nil {
		t.Fatalf("解析 protobuf 失败: %v", err)
	}

	// 转换为 JSON 格式并打印
	jsonData, err := protojson.Marshal(&resp)
	if err != nil {
		t.Fatalf("转换为 JSON 失败: %v", err)
	}

	fmt.Printf("解析后的响应数据:\n%s\n", string(jsonData))

	// 打印各个字段
	fmt.Printf("\n各字段值:\n")
	fmt.Printf("Token: %s\n", resp.Token)
	fmt.Printf("Nickname: %s\n", resp.Nickname)
	fmt.Printf("Email: %s\n", resp.Email)
	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
	}
}
