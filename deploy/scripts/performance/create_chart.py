import pandas as pd
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import seaborn as sns

# -----------------------------
# Configura il file CSV
# -----------------------------
csv_file = "analyze_failure.csv"  # cambia con il tuo file CSV
df = pd.read_csv(csv_file, parse_dates=['timestamp'])

# -----------------------------
# Rinominare colonne con nomi più leggibili
# -----------------------------
df_plot = df.rename(columns={
    'missing_rate_g': 'Missing rate',
    'missing_rate_adj': 'Missing rate - miss',
    'missing_rate_g_sim': 'Missing rate osservato dai simulatori',
    'missing_rate_adj_sim': 'Missing rate simulatore - miss',

    'sim_outlier': 'Outlier generati',
    'outliers': 'Outlier rilevati',
    'outlier_error_percent': 'Errore rilevamento outlier'
})

# -----------------------------
# Grafico 1: Missing rate
# -----------------------------
missing_columns = [
    'Missing rate',
    'Missing rate - miss',
    'Missing rate osservato dai simulatori',
    'Missing rate simulatore - miss'
]

plt.figure(figsize=(8, 10))
sns.boxplot(data=df_plot[missing_columns])
plt.title("Distribuzione delle metriche di Missing Rate")
plt.ylabel("Percentuale (%)")

# Minimo 0, massimo suggerito 20
y_max_suggested = 20
current_max = df_plot[missing_columns].max().max()
plt.ylim(0, max(y_max_suggested, current_max))

# Linea rossa a 20 solo se superata
if current_max > 20:
    plt.axhline(20, color='red', linestyle='--', linewidth=1, label='20%')

plt.xticks(rotation=15)
plt.grid(axis='y', linestyle='--', alpha=0.7)

if current_max > 20:
    plt.legend()

plt.tight_layout()
plt.savefig("missing_rate_boxplot.png")

# -----------------------------
# Grafico 2: Percentuale errore Outlier
# -----------------------------
outlier_percent_columns = ['Errore rilevamento outlier']

plt.figure(figsize=(6, 10))
sns.boxplot(data=df_plot[outlier_percent_columns])
plt.title("Distribuzione Percentuale Errore Rilevamento Outlier")
plt.ylabel("Percentuale (%)")

y_max_suggested = 100
current_max = df_plot[outlier_percent_columns].max().max()
plt.ylim(0, max(y_max_suggested, current_max))

# Linea rossa a 100 solo se superata
if current_max > 100:
    plt.axhline(100, color='red', linestyle='--', linewidth=1, label='100%')

plt.xticks(rotation=15)
plt.grid(axis='y', linestyle='--', alpha=0.7)

if current_max > 100:
    plt.legend()

plt.tight_layout()
plt.savefig("outlier_error_percent_boxplot.png")

# -----------------------------
# Grafico 3: Outlier generati e rilevati (valori assoluti)
# -----------------------------
outlier_abs_columns = ['Outlier generati', 'Outlier rilevati']

plt.figure(figsize=(6, 10))
sns.boxplot(data=df_plot[outlier_abs_columns])
plt.title("Outlier generati vs rilevati (valori assoluti)")
plt.ylabel("Conteggio")

plt.xticks(rotation=15)
plt.grid(axis='y', linestyle='--', alpha=0.7)

plt.tight_layout()
plt.savefig("outlier_abs_boxplot.png")

print(">>> Grafici generati:")
print(" - missing_rate_boxplot.png")
print(" - outlier_error_percent_boxplot.png")
print(" - outlier_abs_boxplot.png")

# -----------------------------
# Configura il file CSV
# -----------------------------
csv_file = "analyze_throughput.csv"
df = pd.read_csv(csv_file, parse_dates=['timestamp'])

# -----------------------------
# Grafico 1: Throughput (boxplot)
# -----------------------------
plt.figure(figsize=(6, 10))
sns.boxplot(y='throughput_msg_per_min', data=df)
plt.title("Distribuzione Throughput Hub (msg/min)")
plt.ylabel("Messaggi per minuto")
plt.grid(axis='y', linestyle='--', alpha=0.7)

# Imposta range y più stretto
# plt.ylim(7, 20)

plt.tight_layout()
plt.savefig("throughput_hub_boxplot.png")
plt.close()


# -----------------------------
# Grafico 2: Latenza end-to-end combinata (minuti)
# -----------------------------
plt.figure(figsize=(6, 10))
lat_min_cols = ['lat_avg_min', 'lat_max_min']
df_melt_min = df.melt(value_vars=lat_min_cols, var_name='Tipo', value_name='Latenza (min)')
sns.boxplot(x='Tipo', y='Latenza (min)', data=df_melt_min, hue='Tipo', dodge=False, palette=['green', 'red'])
plt.title("Distribuzione Latenza end-to-end Hub (minuti)")
plt.grid(axis='y', linestyle='--', alpha=0.7)
plt.tight_layout()
plt.savefig("latency_hub_min_boxplot.png")
plt.close()

# -----------------------------
# Grafico 3: Latenza end-to-end combinata (secondi)
# -----------------------------
plt.figure(figsize=(6, 10))
lat_sec_cols = ['lat_avg_s', 'lat_max_s']
df_melt_sec = df.melt(value_vars=lat_sec_cols, var_name='Tipo', value_name='Latenza (s)')
sns.boxplot(x='Tipo', y='Latenza (s)', data=df_melt_sec, hue='Tipo', dodge=False, palette=['green', 'red'])
plt.title("Distribuzione Latenza end-to-end Hub (secondi)")
plt.grid(axis='y', linestyle='--', alpha=0.7)
plt.tight_layout()
plt.savefig("latency_hub_s_boxplot.png")
plt.close()

print(">>> Grafici generati:")
print(" - throughput_hub_boxplot.png")
print(" - latency_hub_min_boxplot.png")
print(" - latency_hub_s_boxplot.png")