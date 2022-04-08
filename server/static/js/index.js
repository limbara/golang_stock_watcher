/**
 * @param {Array.<Stock>} stocks
 */
const updateTable = (stocks) => {
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

    anime({
      targets: tableBody,
      translateY: 200,
      easing: 'easeInElastic(1, .6)',
      direction: 'reverse',
    });
  };

  const tableBody = document.querySelector('#stock-table tbody');

  if (tableBody.childElementCount != 0) {
    anime({
      targets: tableBody,
      translateY: 200,
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
  console.log(search);
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

window.addEventListener('DOMContentLoaded', async (e) => {
  const stocks = await fetchStocks();
  updateTable(stocks);
});

const debouncedSearchSubmit = debounce(async (e) => {
  const formData = new FormData(e.target);
  const formProps = Object.fromEntries(formData);

  const stocks = await fetchStocks(formProps?.search || '');
  updateTable(stocks);
}, 300);

document.getElementById('form-search').addEventListener('submit', (e) => {
  e.preventDefault();
  debouncedSearchSubmit(e);
});
