# Orbz sGTM Enricher

Serviço interno para enriquecer requisições do Google Tag Manager Server-Side
antes de elas chegarem ao container. Ele é inspirado nos recursos **GEO Headers**
e **User ID** da Stape, mas é uma implementação própria da OrbzTech.

## O que ele retorna

| Cabeçalho | Origem | Uso |
| --- | --- | --- |
| `X-Orbz-User-Id` | HMAC-SHA256 de IP real + User-Agent + host + segredo | identificador pseudônimo para eventos de navegador |
| `X-GEO-Country` | GeoLite2 City local | país ISO, por exemplo `BR` |
| `X-GEO-Region` | GeoLite2 City local | UF/região, por exemplo `CE` |
| `X-GEO-City` | GeoLite2 City local | cidade, quando disponível |
| `X-GEO-PostalCode` | GeoLite2 City local | CEP, quando disponível |

O serviço **não** cria nem substitui `sck`. O `sck` permanece responsável por
ligar a visita, o checkout e o webhook de compra.

## Arquitetura

`Navegador -> Traefik (ForwardAuth) -> GTM Server`

O Orbz Enricher é chamado internamente pelo Traefik. Em uma resposta `204`, o
Traefik copia os cabeçalhos listados e envia a requisição original ao GTM Server.
Não existe rota pública para o enriquecedor e ele não funciona como proxy.

## Segurança

- Use Docker Secrets para o HMAC e para as credenciais MaxMind.
- Não registre IP, User-Agent, segredo ou identificador nos logs.
- Permita acesso ao serviço somente pela rede interna do Traefik.
- Só habilite `trustForwardHeader=true` depois de configurar proxies confiáveis
  nos entrypoints do Traefik. Caso contrário, um visitante pode forjar
  `X-Forwarded-For`.
- Geo é uma aproximação baseada no IP; cidade e CEP podem ficar vazios ou imprecisos.
- Não use GEO/IP/User-Agent de webhooks da Ticto para identificar o comprador:
  eles pertencem ao emissor do webhook.

## Desenvolvimento

O serviço espera a chave HMAC em `/run/secrets/orbz_user_id_hmac` e, quando
presente, a base em `/data/GeoLite2-City.mmdb`.

```bash
go test ./...
docker build -t orbz-sgtm-enricher:local .
```

## Implantação

1. Crie no Swarm os secrets `orbz_user_id_hmac`, `maxmind_account_id` e
   `maxmind_license_key`. Nunca coloque seus valores na stack ou no Git.
2. Configure `ORBZ_ENRICHER_IMAGE` e `TRAEFIK_NETWORK` no Portainer.
3. Faça o deploy de `stack/orbz-enricher.example.yml` como stack própria.
4. Acrescente os labels de `traefik/forwardauth.labels.example.yml` ao GTM Server
   e ao proxy Caddy de debug, ajustando os nomes de stack e roteador reais.
5. No Preview do GTM, crie variáveis de Cabeçalho da Solicitação para cada
   `X-Orbz-*` / `X-GEO-*` e valide primeiro um PageView.

O workflow publica automaticamente no GHCR em cada push na `main`, incluindo
uma imagem separada do atualizador GeoLite2 que lê as credenciais como arquivos
de Docker Secret.
