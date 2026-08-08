# Desafio Cotacao (Client x Server)

Projeto do desafio de Go com:

- HTTP
- Context com timeout
- SQLite
- Manipulacao de arquivo

## Estrutura

- `server/server.go`: sobe servidor HTTP na porta `8080` com endpoint `/cotacao`
- `client/client.go`: consome o endpoint local e salva a cotacao em `cotacao.txt`

## Requisitos

- Go instalado (versao 1.20+ recomendada)

## Como rodar

No terminal, entre na pasta do desafio:

```bash
cd Desafios/01-cotacao
```

Baixe as dependencias:

```bash
go mod tidy
```

### 1) Rodar o servidor

```bash
go run ./server
```

O servidor ficara ouvindo em:

`http://localhost:8080/cotacao`

### 2) Rodar o cliente

Em outro terminal, na mesma pasta:

```bash
go run ./client
```

Se der tudo certo, sera criado o arquivo:

- `cotacao.txt`

Com o formato:

- `Dólar: {valor}`

Exemplo:

- `Dólar: 5.08059`

## API Awesome

De acordo com o enunciado do desafio, a url informada para consultar a cotação está retornando 429: Quota exceeded  
Para retornar a cotação será necessário se cadastrar no site, obter o token e passá-lo como parâmetro na url conforme documentação da API. 


## Regras de timeout implementadas

- Server -> API externa: `200ms`
- Server -> Banco SQLite: `10ms`
- Client -> Server local: `300ms`

Quando algum timeout estoura, o erro e logado no console conforme solicitado.
