// SURFGO_ALPHA_v12.5: FULL_STATEFUL_ALGO
export default {
  async scheduled(event, env, ctx) {
    ctx.waitUntil(this.executeStrategy(env));
  },

  async fetch(request, env) {
    const { pathname } = new URL(request.url);
    if (pathname === "/run-manual") return new Response(await this.executeStrategy(env));
    // Standard HUD/Blog logic remains below...
    return new Response("ALGO_LIVE_MONITORING", { status: 200 });
  },

  async executeStrategy(env) {
    const symbol = "BTC/USD";
    const auth = { headers: { "APCA-API-KEY-ID": env.ALPACA_KEY, "APCA-API-SECRET-KEY": env.ALPACA_SECRET }};

    // 1. DATA ACQUISITION (1-Min Bars for 30 periods)
    const data = await fetch(`https://data.alpaca.markets/v1beta3/crypto/us/bars?symbols=${symbol}&timeframe=1Min&limit=30`, auth).then(r => r.json());
    const bars = data.bars[symbol].map(b => b.c);
    const currentPrice = bars[bars.length - 1];

    // 2. MATH ENGINE: INDICATOR CALCULATION
    const getEMA = (data, period) => {
      const k = 2 / (period + 1);
      return data.reduce((acc, val) => val * k + acc * (1 - k));
    };

    const ema5  = getEMA(bars.slice(-5), 5);
    const ema20 = getEMA(bars.slice(-20), 20);
    
    // RSI Calculation (Simplified 14-period)
    let gains = 0, losses = 0;
    for (let i = bars.length - 14; i < bars.length; i++) {
      let diff = bars[i] - bars[i-1];
      diff > 0 ? gains += diff : losses -= diff;
    }
    const rsi = 100 - (100 / (1 + (gains / (losses || 1))));

    // 3. RETRIEVE RECALL (KV Memory)
    const state = JSON.parse(await env.WAITING_LIST.get("ALGO_STATE") || '{"pos":false,"last_price":0}');

    // 4. DECISION MATRIX
    let signal = "NEUTRAL";
    // BUY: EMA Cross Up + Not Overbought (RSI < 70)
    if (ema5 > ema20 && rsi < 70 && !state.pos) {
      signal = "STRONG_BUY";
      const account = await fetch("https://paper-api.alpaca.markets/v2/account", auth).then(r => r.json());
      const qty = (parseFloat(account.buying_power) * 0.10) / currentPrice; // 10% risk
      await fetch("https://paper-api.alpaca.markets/v2/orders", {
        method: "POST", ...auth,
        body: JSON.stringify({ symbol, qty: qty.toFixed(5), side: "buy", type: "market", time_in_force: "gtc" })
      });
      state.pos = true;
    } 
    // SELL: EMA Cross Down OR Overbought (RSI > 80)
    else if ((ema5 < ema20 || rsi > 80) && state.pos) {
      signal = "EXIT_POSITION";
      await fetch(`https://paper-api.alpaca.markets/v2/positions/${symbol.replace('/','')}`, { method: "DELETE", ...auth });
      state.pos = false;
    }

    // 5. COMMIT TO MEMORY
    state.last_price = currentPrice;
    await env.WAITING_LIST.put("ALGO_STATE", JSON.stringify(state));

    return `PULSE: BTC @ ${currentPrice} | RSI: ${rsi.toFixed(2)} | EMA5/20: ${ema5.toFixed(2)}/${ema20.toFixed(2)} | SIGNAL: ${signal}`;
  }
};
