import { useState, useEffect } from 'react';
import { 
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer 
} from 'recharts';

const getStatusColor = (level) => {
  if (level === 'ERROR') return '#ef4444';   
  if (level === 'WARNING') return '#f59e0b'; 
  return '#10b981';                          
};

// Grafikler için sade, tek ve şık kurumsal renkler
const getChartColor = (type) => {
  if (type === 'cpu') return '#3b82f6';  // Mavi
  if (type === 'ram') return '#8b5cf6';  // Mor
  if (type === 'disk') return '#10b981'; // Yeşil
  return '#38bdf8';
};

const ProgressBar = ({ label, metric }) => {
  if (!metric) return (
    <div style={{ width: '120px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', marginBottom: '4px' }}>
        <span style={{ color: '#94a3b8' }}>{label}</span><span style={{ color: '#64748b' }}>-%</span>
      </div>
      <div style={{ height: '8px', background: '#334155', borderRadius: '4px' }}></div>
    </div>
  );
  const color = getStatusColor(metric.level);
  const value = Math.min(Math.max(metric.metric_value, 0), 100);
  return (
    <div style={{ width: '120px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', marginBottom: '4px' }}>
        <span style={{ color: '#94a3b8' }}>{label}</span><span style={{ color: '#f8fafc', fontWeight: 'bold' }}>{value.toFixed(0)}%</span>
      </div>
      <div style={{ height: '8px', background: '#334155', borderRadius: '4px', overflow: 'hidden' }}>
        <div style={{ height: '100%', width: `${value}%`, background: color, transition: 'width 0.5s ease-in-out, background-color 0.5s' }}></div>
      </div>
    </div>
  );
};

const MetricCard = ({ title, unit, latest }) => {
  if (!latest) return null;
  const color = getStatusColor(latest.level);
  const measureTime = new Date(latest.timestamp).toLocaleTimeString();
  return (
    <div style={{ background: '#1e293b', border: `1px solid ${color}60`, borderRadius: '12px', padding: '20px', display: 'flex', flexDirection: 'column', width: '240px', flexShrink: 0, boxShadow: `0 4px 20px -2px ${color}30`, transition: 'all 0.3s ease' }}>
      <div>
        <h3 style={{ margin: '0 0 10px 0', color: '#94a3b8', fontSize: '0.95rem', textTransform: 'uppercase' }}>{title}</h3>
        <div><span style={{ background: `${color}20`, color: color, padding: '4px 8px', borderRadius: '6px', fontSize: '0.75rem', fontWeight: 'bold' }}>{latest.level}</span></div>
      </div>
      <div style={{ flexGrow: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '15px 0' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', color: color }}>
          <span style={{ fontSize: '3rem', fontWeight: 'bold', lineHeight: '1', textShadow: `0 0 15px ${color}50` }}>{latest.metric_value.toFixed(1)}</span>
          <span style={{ fontSize: '1.2rem', color: '#64748b', marginLeft: '6px', fontWeight: '500', textShadow: 'none' }}>{unit}</span>
        </div>
      </div>
      <div style={{ fontSize: '0.75rem', color: '#64748b', textAlign: 'center' }}>Son ölçüm: {measureTime}</div>
    </div>
  );
};

const MetricChart = ({ title, type, data, timeRange }) => {
  if (!data || data.length === 0) return (
    <div style={{ background: '#1e293b', borderRadius: '12px', padding: '15px', flexGrow: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><p style={{ color: '#64748b' }}>Bu zaman aralığında veri bulunmuyor.</p></div>
  );

  const maxTime = data[data.length - 1].timestamp; 
  const minTime = maxTime - (timeRange * 60 * 1000);
  const chartColor = getChartColor(type);

  return (
    <div style={{ background: '#1e293b', borderRadius: '12px', padding: '15px', flexGrow: 1, minWidth: 0 }}>
      <h4 style={{ margin: '0 0 15px 0', color: '#f8fafc', fontSize: '0.9rem' }}>
        {title} Trendi (Son {timeRange >= 60 ? `${timeRange / 60} Saat` : `${timeRange} Dakika`})
      </h4>
      <div style={{ height: '180px', width: '100%' }}>
        <ResponsiveContainer>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis dataKey="timestamp" type="number" domain={[minTime, maxTime]} allowDataOverflow={true} tickFormatter={(unixTime) => {
                const formatOptions = { hour: '2-digit', minute: '2-digit' };
                if (timeRange === 1) formatOptions.second = '2-digit';
                return new Date(unixTime).toLocaleTimeString([], formatOptions);
              }} stroke="#64748b" tick={{fontSize: 11}} />
            <YAxis stroke="#64748b" tick={{fontSize: 11}} domain={[0, 100]} />
            <Tooltip labelFormatter={(label) => new Date(label).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })} contentStyle={{ background: '#0f172a', border: '1px solid #334155', borderRadius: '8px', color: '#fff' }} />
            <Line type="monotone" dataKey="value" stroke={chartColor} strokeWidth={3} dot={false} isAnimationActive={false} />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

function App() {
  const [fleetData, setFleetData] = useState([]);
  const [selectedHost, setSelectedHost] = useState(null);
  const [hostEvents, setHostEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [timeRange, setTimeRange] = useState(5);

  const [agentConfig, setAgentConfig] = useState(null);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  // 1. FİLO VERİSİNİ 5 SANİYEDE BİR SESSİZCE YENİLE
  useEffect(() => {
    const fetchFleetData = async () => {
      try {
        const hostsRes = await fetch('http://localhost:8081/api/hosts');
        if (!hostsRes.ok) return;
        const hostIds = await hostsRes.json();

        const fleetPromises = hostIds.map(async (id) => {
          const eventsRes = await fetch(`http://localhost:8081/api/events?agent_id=${id}`);
          const events = await eventsRes.json();
          
          if (events && Array.isArray(events)) {
            events.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
          }

          const getLatest = (type) => events?.find(e => e.check_type === type) || null;
          let isOnline = false;
          if (events && events.length > 0) {
            const latestTime = new Date(events[0].timestamp).getTime();
            const fiveMinsAgo = new Date().getTime() - (5 * 60 * 1000);
            isOnline = latestTime > fiveMinsAgo;
          }
          return { id, isOnline, cpu: getLatest('cpu'), ram: getLatest('ram'), disk: getLatest('disk') };
        });

        const results = await Promise.all(fleetPromises);
        results.sort((a, b) => (a.isOnline === b.isOnline) ? 0 : a.isOnline ? -1 : 1);
        setFleetData(results);
        setLoading(false);
      } catch (err) {
        console.error("Filo verisi çekilirken hata:", err);
      }
    };

    fetchFleetData(); 
    const interval = setInterval(fetchFleetData, 5000); 
    return () => clearInterval(interval); 
  }, []);

  // 2. SEÇİLİ SUNUCUNUN LOGLARINI VE AYARLARINI HER 5 SANİYEDE BİR AKICI OLARAK YENİLE
  useEffect(() => {
    if (!selectedHost) return;

    const fetchConfig = () => {
      fetch(`http://localhost:8081/api/configs/${selectedHost}`)
        .then(res => res.json())
        .then(data => {
          if (!isSettingsOpen) {
            if (data && (data.cpu_warn || data.cpu_err)) {
              setAgentConfig(data);
            } else {
              setAgentConfig({
                agent_id: selectedHost,
                cpu_warn: 75, cpu_err: 90,
                ram_warn: 80, ram_err: 95,
                disk_warn: 85, disk_err: 95
              });
            }
          }
        })
        .catch(err => {
          console.error("Ayarlar çekilemedi, varsayılanlar atanıyor", err);
          if (!isSettingsOpen) {
            setAgentConfig({
              agent_id: selectedHost,
              cpu_warn: 75, cpu_err: 90,
              ram_warn: 80, ram_err: 95,
              disk_warn: 85, disk_err: 95
            });
          }
        });
    };

    const fetchEvents = () => {
      fetch(`http://localhost:8081/api/events?agent_id=${selectedHost}`)
        .then(res => res.json())
        .then(data => {
          if (data && Array.isArray(data)) {
            data.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
          }
          setHostEvents(data || []);
          setLoading(false);
        })
        .catch(err => console.error("Eventler çekilirken hata:", err));
    };

    setLoading(true); 
    fetchConfig();
    fetchEvents();

    const interval = setInterval(() => {
      fetchEvents();
      fetchConfig();
    }, 5000); 

    return () => clearInterval(interval); 
  }, [selectedHost, isSettingsOpen]);

  const getChartData = (type) => {
    const events = hostEvents.filter(e => e.check_type === type);
    if (events.length === 0) return [];
    const latestTime = new Date(events[0].timestamp).getTime(); 
    const cutoffTime = latestTime - (timeRange * 60 * 1000);
    return events
      .filter(e => new Date(e.timestamp).getTime() >= cutoffTime)
      .reverse()
      .map(e => ({ timestamp: new Date(e.timestamp).getTime(), value: e.metric_value, level: e.level }));
  };

  const getLatestMetric = (type) => {
    const events = hostEvents.filter(e => e.check_type === type);
    return events.length > 0 ? events[0] : null; 
  };

  const saveConfig = async () => {
    const requiredKeys = ['cpu_warn', 'cpu_err', 'ram_warn', 'ram_err', 'disk_warn', 'disk_err'];
    for (const key of requiredKeys) {
      if (agentConfig[key] === "" || agentConfig[key] === null || agentConfig[key] === undefined) {
        alert("Lütfen tüm eşik değerlerini doldurunuz! (Alanlar boş bırakılamaz)");
        return;
      }
    }

    setIsSaving(true);
    try {
      await fetch(`http://localhost:8081/api/configs/${selectedHost}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...agentConfig,
          cpu_warn: Number(agentConfig.cpu_warn),
          cpu_err: Number(agentConfig.cpu_err),
          ram_warn: Number(agentConfig.ram_warn),
          ram_err: Number(agentConfig.ram_err),
          disk_warn: Number(agentConfig.disk_warn),
          disk_err: Number(agentConfig.disk_err),
        })
      });
      setIsSettingsOpen(false); 
    } catch (error) {
      alert("Kaydedilirken hata oluştu!");
    }
    setIsSaving(false);
  };

  // Ortak input stili: SOLA DAYALI, temiz ve ferah tasarım
  const inputStyle = {
    background: '#1e293b', 
    color: '#fff', 
    border: '1px solid #475569', 
    padding: '10px 14px', 
    borderRadius: '8px',
    textAlign: 'left',
    fontFamily: 'Inter, system-ui, sans-serif',
    fontSize: '0.95rem',
    outline: 'none',
    width: '100%',
    boxSizing: 'border-box'
  };

  // Sadece sayı girilmesine izin veren ve değiştiren yardımcı fonksiyon
  const handleNumberChange = (key, value) => {
    if (value === '' || /^[0-9]+$/.test(value)) {
      setAgentConfig({ ...agentConfig, [key]: value });
    }
  };

  // Tıklandığında imleci kesin olarak metnin en sonuna (sağına) götüren fonksiyon
  const moveCursorToEnd = (e) => {
    const len = e.target.value.length;
    e.target.setSelectionRange(len, len);
  };

  return (
    <div style={{ backgroundColor: '#0f172a', minHeight: '100vh', color: '#e2e8f0', fontFamily: 'Inter, system-ui, sans-serif', padding: '40px' }}>
      <div style={{ maxWidth: '1400px', margin: '0 auto' }}>
        
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '40px' }}>
          <div>
            <h1 style={{ margin: 0, fontSize: '2rem', color: '#f8fafc' }}>PulseGuard</h1>
            <p style={{ margin: '5px 0 0 0', color: '#94a3b8' }}>Kurumsal Filo İzleme Merkezi</p>
          </div>
          {selectedHost && (
            <button 
              onClick={() => { setSelectedHost(null); setIsSettingsOpen(false); }}
              style={{ background: '#334155', color: '#fff', border: 'none', padding: '10px 20px', borderRadius: '8px', cursor: 'pointer', fontWeight: 'bold', transition: '0.2s' }}
              onMouseOver={e => e.target.style.background = '#475569'}
              onMouseOut={e => e.target.style.background = '#334155'}
            >
              ⬅ Filo Görünümüne Dön
            </button>
          )}
        </div>

        {!selectedHost && (
          <div>
            <h2 style={{ fontSize: '1.25rem', color: '#e2e8f0', marginBottom: '20px' }}>Bağlı Sunucular (Canlı)</h2>
            {loading && fleetData.length === 0 && <p style={{ color: '#94a3b8' }}>Sistem taranıyor...</p>}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '15px' }}>
              {fleetData.map((host, idx) => (
                <div 
                  key={idx} onClick={() => setSelectedHost(host.id)}
                  style={{ background: '#1e293b', padding: '20px', borderRadius: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderLeft: `4px solid ${host.isOnline ? '#10b981' : '#475569'}`, transition: 'transform 0.2s, background 0.2s', opacity: host.isOnline ? 1 : 0.6 }}
                  onMouseOver={e => { e.currentTarget.style.transform = 'translateY(-2px)'; e.currentTarget.style.background = '#334155'; }}
                  onMouseOut={e => { e.currentTarget.style.transform = 'translateY(0)'; e.currentTarget.style.background = '#1e293b'; }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '15px', width: '200px' }}>
                    <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: host.isOnline ? '#10b981' : '#475569', boxShadow: host.isOnline ? '0 0 10px #10b981' : 'none' }}></div>
                    <span style={{ fontSize: '1.1rem', fontWeight: '600', color: host.isOnline ? '#f8fafc' : '#94a3b8' }}>{host.id}</span>
                  </div>
                  <div style={{ display: 'flex', gap: '40px', flexGrow: 1, justifyContent: 'center' }}>
                    <ProgressBar label="CPU" metric={host.cpu} />
                    <ProgressBar label="RAM" metric={host.ram} />
                    <ProgressBar label="DİSK" metric={host.disk} />
                  </div>
                  <div style={{ width: '50px', textAlign: 'right' }}><span style={{ color: '#64748b' }}>➔</span></div>
                </div>
              ))}
            </div>
          </div>
        )}

        {selectedHost && (
          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '30px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '15px' }}>
                <div style={{ width: '16px', height: '16px', borderRadius: '50%', background: '#10b981', boxShadow: '0 0 12px #10b981' }}></div>
                <h2 style={{ fontSize: '1.5rem', color: '#f8fafc', margin: 0 }}>{selectedHost} - Canlı Akış</h2>
              </div>
              
              <div style={{ display: 'flex', alignItems: 'center', gap: '15px' }}>
                <button 
                  onClick={() => setIsSettingsOpen(!isSettingsOpen)}
                  style={{ background: isSettingsOpen ? '#3b82f6' : '#1e293b', color: '#fff', border: '1px solid #3b82f6', padding: '8px 16px', borderRadius: '8px', cursor: 'pointer', fontSize: '0.9rem', fontWeight: 'bold', transition: '0.2s' }}
                >
                  ⚙️ Eşikleri Düzenle
                </button>

                <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                  <label style={{ color: '#94a3b8', fontSize: '0.9rem' }}>Zaman Aralığı:</label>
                  <select value={timeRange} onChange={(e) => setTimeRange(Number(e.target.value))} style={{ background: '#1e293b', color: '#f8fafc', border: '1px solid #334155', padding: '8px 16px', borderRadius: '8px', cursor: 'pointer', fontSize: '0.9rem', outline: 'none' }}>
                    <option value={1}>Son 1 Dakika</option>
                    <option value={5}>Son 5 Dakika</option>
                    <option value={15}>Son 15 Dakika</option>
                    <option value={60}>Son 1 Saat</option>
                  </select>
                </div>
              </div>
            </div>

            {isSettingsOpen && (
              <div style={{ background: '#1e293b', padding: '25px', borderRadius: '12px', marginBottom: '25px', border: '1px solid #334155' }}>
                <h3 style={{ margin: '0 0 20px 0', color: '#f8fafc' }}>🚨 {selectedHost} Eşik Değerleri (Thresholds)</h3>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '20px' }}>
                  
                  {/* CPU EŞİKLERİ */}
                  <div style={{ background: '#0f172a', padding: '15px', borderRadius: '8px' }}>
                    <h4 style={{ margin: '0 0 12px 0', color: '#3b82f6', fontSize: '1rem' }}>CPU Eşikleri</h4>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                      <label style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Uyarı Sınırı (Warning %)</label>
                      <input 
                        type="text" 
                        inputMode="numeric"
                        value={agentConfig?.cpu_warn ?? ""} 
                        onChange={(e) => handleNumberChange('cpu_warn', e.target.value)} 
                        onFocus={moveCursorToEnd}
                        onClick={moveCursorToEnd}
                        style={{ ...inputStyle, borderColor: '#f59e0b' }} 
                      />
                      <label style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Kritik Sınır (Error %)</label>
                      <input 
                        type="text" 
                        inputMode="numeric"
                        value={agentConfig?.cpu_err ?? ""} 
                        onChange={(e) => handleNumberChange('cpu_err', e.target.value)} 
                        onFocus={moveCursorToEnd}
                        onClick={moveCursorToEnd}
                        style={{ ...inputStyle, borderColor: '#ef4444' }} 
                      />
                    </div>
                  </div>

                  {/* RAM EŞİKLERİ */}
                  <div style={{ background: '#0f172a', padding: '15px', borderRadius: '8px' }}>
                    <h4 style={{ margin: '0 0 12px 0', color: '#8b5cf6', fontSize: '1rem' }}>RAM Eşikleri</h4>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                      <label style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Uyarı Sınırı (Warning %)</label>
                      <input 
                        type="text" 
                        inputMode="numeric"
                        value={agentConfig?.ram_warn ?? ""} 
                        onChange={(e) => handleNumberChange('ram_warn', e.target.value)} 
                        onFocus={moveCursorToEnd}
                        onClick={moveCursorToEnd}
                        style={{ ...inputStyle, borderColor: '#f59e0b' }} 
                      />
                      <label style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Kritik Sınır (Error %)</label>
                      <input 
                        type="text" 
                        inputMode="numeric"
                        value={agentConfig?.ram_err ?? ""} 
                        onChange={(e) => handleNumberChange('ram_err', e.target.value)} 
                        onFocus={moveCursorToEnd}
                        onClick={moveCursorToEnd}
                        style={{ ...inputStyle, borderColor: '#ef4444' }} 
                      />
                    </div>
                  </div>

                  {/* DİSK EŞİKLERİ */}
                  <div style={{ background: '#0f172a', padding: '15px', borderRadius: '8px' }}>
                    <h4 style={{ margin: '0 0 12px 0', color: '#10b981', fontSize: '1rem' }}>Disk Eşikleri</h4>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                      <label style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Uyarı Sınırı (Warning %)</label>
                      <input 
                        type="text" 
                        inputMode="numeric"
                        value={agentConfig?.disk_warn ?? ""} 
                        onChange={(e) => handleNumberChange('disk_warn', e.target.value)} 
                        onFocus={moveCursorToEnd}
                        onClick={moveCursorToEnd}
                        style={{ ...inputStyle, borderColor: '#f59e0b' }} 
                      />
                      <label style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Kritik Sınır (Error %)</label>
                      <input 
                        type="text" 
                        inputMode="numeric"
                        value={agentConfig?.disk_err ?? ""} 
                        onChange={(e) => handleNumberChange('disk_err', e.target.value)} 
                        onFocus={moveCursorToEnd}
                        onClick={moveCursorToEnd}
                        style={{ ...inputStyle, borderColor: '#ef4444' }} 
                      />
                    </div>
                  </div>

                </div>
                
                {/* İPTAL VE KAYDET BUTONLARI */}
                <div style={{ marginTop: '20px', display: 'flex', justifyContent: 'flex-end', gap: '12px' }}>
                  <button 
                    onClick={() => setIsSettingsOpen(false)} 
                    style={{ background: '#334155', color: '#e2e8f0', border: 'none', padding: '10px 20px', borderRadius: '8px', cursor: 'pointer', fontWeight: 'bold', transition: '0.2s' }}
                    onMouseOver={e => e.target.style.background = '#475569'}
                    onMouseOut={e => e.target.style.background = '#334155'}
                  >
                    ❌ İptal
                  </button>
                  <button 
                    onClick={saveConfig} 
                    disabled={isSaving} 
                    style={{ background: '#10b981', color: '#fff', border: 'none', padding: '10px 20px', borderRadius: '8px', cursor: 'pointer', fontWeight: 'bold' }}
                  >
                    {isSaving ? "Kaydediliyor..." : "💾 Ayarları Kaydet"}
                  </button>
                </div>

              </div>
            )}
            
            {loading ? (
              <p style={{ color: '#94a3b8' }}>Metrikler yükleniyor...</p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '25px' }}>
                <div style={{ display: 'flex', gap: '20px', alignItems: 'stretch' }}>
                  <MetricCard title="CPU Kullanımı" unit="%" latest={getLatestMetric('cpu')} />
                  <MetricChart title="CPU" type="cpu" data={getChartData('cpu')} timeRange={timeRange} />
                </div>
                <div style={{ display: 'flex', gap: '20px', alignItems: 'stretch' }}>
                  <MetricCard title="RAM Kullanımı" unit="%" latest={getLatestMetric('ram')} />
                  <MetricChart title="RAM" type="ram" data={getChartData('ram')} timeRange={timeRange} />
                </div>
                <div style={{ display: 'flex', gap: '20px', alignItems: 'stretch' }}>
                  <MetricCard title="Disk Kullanımı" unit="%" latest={getLatestMetric('disk')} />
                  <MetricChart title="Disk" type="disk" data={getChartData('disk')} timeRange={timeRange} />
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default App;