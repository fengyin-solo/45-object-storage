// 对象存储管理控制台前端脚本，原生 JS 零依赖。
const API_KEY = 'objectstore-secret-key';

async function api(path) {
  const res = await fetch(path, { headers: { 'X-API-Key': API_KEY } });
  const body = await res.json();
  if (body.code !== 0) {
    throw new Error(body.message || '请求失败');
  }
  return body.data;
}

function formatBytes(n) {
  if (n === null || n === undefined) return '-';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  let v = Number(n);
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return v.toFixed(2) + ' ' + units[i];
}

function fmtTime(iso) {
  if (!iso) return '-';
  const d = new Date(iso);
  return d.toLocaleString('zh-CN', { hour12: false });
}

function renderStats(overview) {
  const items = [
    ['存储桶', overview.bucket_count],
    ['对象', overview.object_count],
    ['活跃对象', overview.active_object_count],
    ['总容量', formatBytes(overview.total_bytes)],
    ['分片上传', overview.multipart_upload_count],
    ['生命周期规则', overview.lifecycle_rule_count],
    ['桶策略', overview.policy_count],
    ['访问日志', overview.access_log_count],
  ];
  document.getElementById('stats').innerHTML = items.map(([label, value]) =>
    `<div class="card"><div class="label">${label}</div><div class="value">${value}</div></div>`
  ).join('');
}

function renderBuckets(buckets) {
  const tbody = document.querySelector('#buckets tbody');
  tbody.innerHTML = '';
  (buckets.items || []).forEach(b => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${b.id}</td>
      <td>${b.name}</td>
      <td>${b.region}</td>
      <td>${b.owner}</td>
      <td>${formatBytes(b.quota_bytes)}</td>
      <td><span class="badge ${b.status}">${b.status}</span></td>
      <td>${fmtTime(b.created_at)}</td>`;
    tbody.appendChild(tr);
  });
}

function renderRegions(regions) {
  const tbody = document.querySelector('#regions tbody');
  tbody.innerHTML = '';
  (regions || []).forEach(r => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${r.region}</td>
      <td>${r.bucket_count}</td>
      <td>${r.object_count}</td>
      <td>${formatBytes(r.total_bytes)}</td>`;
    tbody.appendChild(tr);
  });
}

async function load() {
  try {
    const [overview, buckets, regions] = await Promise.all([
      api('/api/stats/overview'),
      api('/api/buckets?size=100'),
      api('/api/stats/by-region'),
    ]);
    renderStats(overview);
    renderBuckets(buckets);
    renderRegions(regions);
  } catch (err) {
    document.getElementById('stats').innerHTML =
      `<div class="card"><div class="label">加载失败</div><div class="value">${err.message}</div></div>`;
  }
}

document.getElementById('refresh').addEventListener('click', load);
load();
