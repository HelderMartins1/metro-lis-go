<script>
  import { site } from '../lib/site.js';
  import Page from './Page.svelte';
</script>

<Page title="Política de privacidade">
  <p>
    Esta política explica que dados o {site.name} trata quando o usas. Não há contas, publicidade, cookies
    de rastreio nem estatísticas de utilização, e nenhum dado é vendido ou partilhado para fins de
    marketing.
  </p>

  <h2>Quem é responsável</h2>
  <p>
    O {site.name} é uma app pessoal, independente e não oficial, sem qualquer ligação ao Metropolitano de
    Lisboa. O responsável pelo projeto é também o responsável pelo tratamento dos dados descritos aqui.
  </p>
  {#if site.contact}
    <p>Para qualquer questão sobre os teus dados, escreve para {site.contact}.</p>
  {/if}

  <h2>Dados guardados no teu dispositivo</h2>
  <p>
    Para a app funcionar como esperas, ficam guardadas no armazenamento local do teu dispositivo as tuas
    preferências (como as estações favoritas e o tema), algumas informações de utilização recente (como a
    última estação aberta e os últimos tempos de espera) e os ficheiros da app, para poder abrir sem rede.
  </p>
  <p>
    Estes dados não saem do dispositivo e podes apagá-los a qualquer momento, limpando os dados do site
    ou desinstalando a app. Base legal: o interesse legítimo em prestar o serviço que pedes (artigo 6.º,
    n.º 1, alínea f) do RGPD).
  </p>

  <h2>Dados que o servidor recebe</h2>
  <p>
    Para mostrar tempos de espera, o estado das linhas e a posição dos comboios, a app comunica com o
    servidor do {site.name}, que por sua vez obtém os dados públicos do Metropolitano de Lisboa. Nessa
    comunicação, o servidor recebe os dados técnicos necessários para responder aos pedidos, como
    informação de ligação e o conteúdo consultado.
  </p>
  <p>
    Estes dados servem apenas para prestar o serviço, mantê-lo a funcionar e protegê-lo de abusos, e são
    guardados só pelo tempo necessário a esses fins. Não são usados para te identificar nem para criar
    perfis. Base legal: o interesse legítimo em prestar e proteger o serviço (artigo 6.º, n.º 1, alínea
    f) do RGPD).
  </p>

  <h2>Notificações</h2>
  <p>
    As notificações estão desligadas até as ligares. Quando as ligas, o servidor guarda o necessário para
    as enviar: um identificador de entrega criado pelo serviço de notificações do teu dispositivo e as
    escolhas que fizeste (estações, lembretes e comboios que segues). As mensagens chegam ao dispositivo
    através desse serviço de notificações.
  </p>
  <p>
    Base legal: o teu consentimento (artigo 6.º, n.º 1, alínea a) do RGPD), que podes retirar a qualquer
    momento desligando as notificações na app; nesse caso, estes dados são apagados.
  </p>

  <h2>Os teus direitos</h2>
  <p>
    Nos termos do RGPD, tens direito de acesso, retificação, apagamento, limitação e oposição ao
    tratamento dos teus dados, e, quando o tratamento se baseia no consentimento, a retirá-lo sem afetar
    o tratamento já feito. Podes também apresentar reclamação à Comissão Nacional de Proteção de Dados
    (CNPD).
  </p>

  <h2>Crianças</h2>
  <p>A app não pede dados pessoais e não se destina a recolher dados de crianças.</p>

  <h2>Alterações</h2>
  <p>
    Se esta política mudar, a nova versão fica publicada nesta página, com a data de atualização abaixo.
  </p>

  <p class="meta">Última atualização: {site.policyDate}.</p>
</Page>
