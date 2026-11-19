// Turns the raw /api/{code}/trains answer into cards grouped by line.
//
// The API only gives a destination id per platform. A destination can belong
// to more than one line (Baixa-Chiado, Alameda, Campo Grande, São Sebastião),
// so the line is the one the destination and this station have in common.

function arrivals(entry) {
  const pairs = [
    [entry.comboio, entry.tempoChegada1],
    [entry.comboio2, entry.tempoChegada2],
    [entry.comboio3, entry.tempoChegada3],
  ];
  return pairs
    .filter(([, t]) => t !== null && t !== undefined && t !== '' && Number.isFinite(Number(t)))
    .map(([train, t], i) => ({ id: String(train || `${entry.cais}-${i}`), seconds: Number(t) }));
}

export function groupBoard(entries, station, lines, destinations) {
  const groups = new Map(station.lines.map((id) => [id, new Map()]));
  const unknown = new Map();

  for (const entry of entries) {
    const destId = String(entry.destino ?? '');
    const dest = destinations?.[destId];
    // Trains that end here are of no use to someone waiting on the platform.
    if (dest && dest.name === station.name) continue;

    let line = dest ? station.lines.find((id) => dest.lines.includes(id)) : undefined;
    if (!line && station.lines.length === 1) line = station.lines[0];

    const bucket = line ? groups.get(line) : unknown;
    const card = bucket.get(destId) ?? { key: destId, dest: dest?.name ?? null, trains: [] };
    card.trains.push(...arrivals(entry));
    bucket.set(destId, card);
  }

  // Two platforms can report the same train; keep each train once.
  const finish = (cards) =>
    [...cards.values()].map((card) => ({
      ...card,
      trains: [...new Map(card.trains.map((t) => [t.id, t])).values()]
        .sort((a, b) => a.seconds - b.seconds)
        .slice(0, 3),
    }));

  return {
    groups: station.lines.map((id) => {
      const order = lines.find((l) => l.id === id)?.stations.map((s) => s.name) ?? [];
      const position = (card) => (card.dest ? order.indexOf(card.dest) : Infinity);
      return { id, cards: finish(groups.get(id)).sort((a, b) => position(a) - position(b)) };
    }),
    unknown: finish(unknown),
  };
}
