/**
 * @param {Array.<Stock>} stocks
 */
const updateTable = (stocks) => {
  const tableBody = document.querySelector('#stock-table tbody');

  const update = () => {
    const tableRows = stocks.map((stock) =>
      createStockTr(stock, (value) => {
        const td = document.createElement('td');
        td.style.opacity = 1;
        td.innerHTML = value;
        return td;
      })
    );

    tableBody.append(...tableRows);

    tableBody.style.opacity = 0;
    anime({
      targets: tableBody,
      translateY: 100,
      easing: 'easeInElastic(1, .6)',
      direction: 'reverse',
      update: (anim) => {
        tableBody.style.opacity = (100 - Math.round(anim.progress)) * 0.01;
      },
    });
  };

  if (tableBody.childElementCount != 0) {
    anime({
      targets: tableBody,
      translateY: 50,
      easing: 'easeInElastic(1, .6)',
      complete: () => {
        while (tableBody.firstChild) {
          tableBody.removeChild(tableBody.lastChild);
        }
        tableBody.style.transform = '';
        update();
      },
    });
  } else {
    update();
  }
};

/**
 * @param {Stock} stock
 * @param {function} createTdFunc function(value) => td element
 */
const createStockTr = (stock, createTdFunc) => {
  const tr = document.createElement('tr');

  const order = ['code', 'name', 'open', 'close', 'high', 'low'];

  order.forEach((key) => {
    const td = createTdFunc(stock[key]);
    tr.appendChild(td);
  });

  return tr;
};

/**
 * @typedef {{"_id": string, "code": string, "name": string, "open": number, "close": number, "high": number, "low": number}} Stock
 * @param {String} search
 * @returns {Array.<Stock>}
 */
const fetchStocks = async (search = '') => {
  const response = await fetch(
    `/api/v1/stocks${search != '' ? `?search=${search}` : ''}`,
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );

  const stocks = await response.json();

  return stocks;
};

/**
 * @param {function} func
 * @param {number} timeout
 */
const debounce = function (func, timeout = 300) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      func.apply(this, args);
    }, timeout);
  };
};

let lastSearchValue;

window.addEventListener('DOMContentLoaded', async (e) => {
  const stocks = await fetchStocks();
  updateTable(stocks);

  document.getElementById('search-input').value = '';
  lastSearchValue = '';
});

const debouncedSearchSubmit = debounce(async (e) => {
  const formData = new FormData(e.target);
  const formProps = Object.fromEntries(formData);

  if (lastSearchValue == formProps?.search || '') {
    return;
  }

  const stocks = await fetchStocks(formProps?.search || '');
  updateTable(stocks);

  lastSearchValue = formProps?.search || '';
}, 200);

document.getElementById('form-search').addEventListener('submit', (e) => {
  e.preventDefault();
  debouncedSearchSubmit(e);
});
