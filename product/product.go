package product

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/integrmais/bling"
	"github.com/integrmais/bling/internal"
)

type ProductService struct {
	AppKey string
	Client *http.Client
}

var Page int

func NewProductService(appKey string, c *http.Client) *ProductService {
	return &ProductService{
		AppKey: appKey,
		Client: c,
	}
}

func HandlerError(req ResponseModel) error {
	if req.Response.Errors.Error.Code == 0 {
		return nil
	}

	return errors.New(
		req.Response.Errors.Error.Message,
	)
}

func HandlerResponse(res *http.Response) (ResponseModel, error) {
	if res.StatusCode == http.StatusBadRequest {
		return ResponseModel{}, errors.New(fmt.Sprintf("Bad Request %d", res.StatusCode))
	}

	if res.StatusCode == http.StatusTooManyRequests {
		return ResponseModel{}, errors.New(fmt.Sprintf("Too Many Requests %d", res.StatusCode))
	}

	var response ResponseModel
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		body, _ := io.ReadAll(res.Body)

		fmt.Println(string(body))

		return ResponseModel{}, err
	}

	err = HandlerError(response)
	if err != nil {
		return ResponseModel{}, err
	}

	return response, nil
}

func (s *ProductService) GetProductById(ctx context.Context, productID string) (ResponseModel, error) {
	url := fmt.Sprintf(
		"%s/produto/%s/%s",
		bling.DefaultUrl,
		productID,
		bling.DefaultResponseType,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ResponseModel{}, err
	}

	q := req.URL.Query()
	q.Add("apikey", s.AppKey)
	q.Add("imagem", "S")
	q.Add("estoque", "S")
	q.Add("tipo", "P")

	req.URL.RawQuery = q.Encode()

	res, err := s.Client.Do(req)
	if err != nil {
		return ResponseModel{}, err
	}

	p, err := HandlerResponse(res)
	if err != nil {
		return ResponseModel{}, err
	}

	return p, nil
}

func (s *ProductService) GetByRange(ctx context.Context, startAt time.Time, page int) (ResponseModel, error) {
	if page <= 0 {
		page = 1
	}

	url := fmt.Sprintf(
		"%s/page=%d/%s",
		bling.ProductsUrl,
		page,
		bling.DefaultResponseType,
	)

	by := internal.NormalizeDate(startAt)
	at := internal.NormalizeDate(time.Now())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ResponseModel{}, err
	}

	q := req.URL.Query()
	q.Add("apikey", s.AppKey)
	q.Add("imagem", "S")
	q.Add("estoque", "S")
	q.Add("tipo", "P")
	q.Add("filters", fmt.Sprintf("dataAlteracao[%s TO %s]", by, at))

	req.URL.RawQuery = q.Encode()

	res, err := s.Client.Do(req)
	if err != nil {
		return ResponseModel{}, err
	}

	p, err := HandlerResponse(res)
	if err != nil {
		return ResponseModel{}, err
	}

	return p, nil
}

func (s *ProductService) Create(ctx context.Context, product Product) (ResponseCreatorModel, error) {
	productXml, err := xml.Marshal(product)
	if err != nil {
		return ResponseCreatorModel{}, err
	}

	fmt.Println(string(productXml))

	url := fmt.Sprintf(
		"%s/produto%s",
		bling.DefaultUrl,
		bling.DefaultResponseType,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return ResponseCreatorModel{}, err
	}

	q := req.URL.Query()
	q.Add("apikey", s.AppKey)
	q.Add("xml", string(productXml))

	req.URL.RawQuery = q.Encode()

	res, err := s.Client.Do(req)
	if err != nil {
		return ResponseCreatorModel{}, err
	}

	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)

	var response ResponseCreatorModel
	if err := json.Unmarshal(b, &response); err != nil {
		fmt.Println(string(b))

		return ResponseCreatorModel{}, err
	}

	return response, nil
}
