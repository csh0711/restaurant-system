import React, { useEffect, useState } from "react";
import "./App.css";

const API_URL = "http://localhost:8080";


function App() {
  const [menu, setMenu] = useState([]);
  const [selectedItems, setSelectedItems] = useState([]);
  const [tableId, setTableId] = useState("");
  const [message, setMessage] = useState(null);

  useEffect(() => {
    fetch(`${API_URL}/menu`)
        .then((res) => res.json())
        .then(setMenu)
        .catch(() => setMessage("Failed to load menu"));
  }, []);

  const toggleItem = (item) => {
    setSelectedItems((prev) =>
        prev.includes(item)
            ? prev.filter((i) => i !== item)
            : [...prev, item]
    );
  };

  const submitOrder = async () => {
    setMessage(null);

    try {
      const res = await fetch(`${API_URL}/orders`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          tableId,
          items: selectedItems,
        }),
      });

      if (!res.ok) throw new Error();

      const data = await res.json();

      setMessage({
        type: "success",
        text: `Order created: ${data.orderId}`,
      });

      setSelectedItems([]);
    } catch {
      setMessage({
        type: "error",
        text: "Failed to create order",
      });
    }
  };

  return (
      <div className="container">
        <h1>🍽 Restaurant Ordering</h1>

        <div>
          <label>Table ID</label>
          <input
              className="input"
              value={tableId}
              onChange={(e) => setTableId(e.target.value)}
              placeholder="Enter table number (e.g. 1)"
          />
        </div>

        <h2>Menu</h2>
        <div className="menu">
          {menu.map((item) => (
              <div className="menu-item" key={item}>
                <input
                    type="checkbox"
                    checked={selectedItems.includes(item)}
                    onChange={() => toggleItem(item)}
                />
                {item}
              </div>
          ))}
        </div>

        <button
            className="button"
            onClick={submitOrder}
            disabled={!tableId || selectedItems.length === 0}
        >
          Place Order
        </button>

        {message && (
            <div className={`message ${message.type} fade-in`}>
              {message.text}
            </div>
        )}
      </div>
  );
}

export default App;