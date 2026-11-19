<script>
  import { site } from '../lib/site.js';
  import Page from './Page.svelte';
</script>

<Page title="Sobre o {site.name}">
  <p class="notice">
    O {site.name} é uma app independente e não oficial. Não tem qualquer ligação ao Metropolitano de
    Lisboa, E.P.E., nem é apoiada ou revista por ele.
  </p>

  <h2>O que faz</h2>
  <p>
    Mostra, estação a estação, quanto falta para os próximos comboios em cada sentido e o estado de
    cada linha. Os tempos atualizam-se sozinhos a cada 20 segundos.
  </p>

  <p>
    Se ligares as notificações, avisa-te quando o comboio que segues está a 2 minutos, quando há
    perturbações nas linhas das tuas estações, e às horas que escolheres com os próximos comboios.
  </p>

  <h2>De onde vêm os dados</h2>
  <p>
    Os tempos de espera, os destinos e o estado das linhas vêm da API pública do Metropolitano de
    Lisboa. O {site.name} só os apresenta: não os cria nem os corrige.
  </p>
  <p>
    Por isso podem ter atrasos, falhas ou erros. Em caso de dúvida, confia nos painéis da estação e
    nos avisos oficiais do Metro.
  </p>

  <h2>Marcas</h2>
  <p>
    “Metropolitano de Lisboa” e “Metro de Lisboa” são marcas do seu titular, usadas aqui só para
    dizer de que rede se trata. Os nomes das linhas e das estações são os nomes públicos da rede.
  </p>

  <h2>Privacidade</h2>
  <p>
    Não há contas, publicidade nem estatísticas de utilização. Lê a
    <a href="#/privacidade">política de privacidade</a> para saberes o pouco que é guardado e onde.
  </p>
  <p class="actions">
    <a class="btn btn--primary" href="#/privacidade">Ler a política de privacidade</a>
  </p>
</Page>
