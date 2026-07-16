package auth

import (
	"context"
	"encoding/json"
	"errors"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysms "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliyunSender struct {
	client                 *dysms.Client
	signName, templateCode string
}

func NewAliyunSender(accessKeyID, accessKeySecret, signName, templateCode string) (*AliyunSender, error) {
	if accessKeyID == "" || accessKeySecret == "" || signName == "" || templateCode == "" {
		return nil, errors.New("aliyun SMS configuration incomplete")
	}
	client, err := dysms.NewClient(&openapi.Config{AccessKeyId: tea.String(accessKeyID), AccessKeySecret: tea.String(accessKeySecret), Endpoint: tea.String("dysmsapi.aliyuncs.com")})
	if err != nil {
		return nil, err
	}
	return &AliyunSender{client: client, signName: signName, templateCode: templateCode}, nil
}
func (s *AliyunSender) SendCode(ctx context.Context, phone, code string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	params, _ := json.Marshal(map[string]string{"code": code})
	request := &dysms.SendSmsRequest{PhoneNumbers: tea.String(phone), SignName: tea.String(s.signName), TemplateCode: tea.String(s.templateCode), TemplateParam: tea.String(string(params))}
	response, err := s.client.SendSms(request)
	if err != nil {
		return err
	}
	if response == nil || response.Body == nil || tea.StringValue(response.Body.Code) != "OK" {
		return errors.New("aliyun SMS rejected request")
	}
	return nil
}
