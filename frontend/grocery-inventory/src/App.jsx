import React, { useState } from "react";

const App = () => {
  const [items, setItems] = useState([{
    upc: "",
    name: "",
    image: "",
    count: 0,
  }]);

  const [formData, setFormData] = useState({
    upc: "",
  });

  // Update form data to always be up to date with the entries on the form.
  const handleChange = (event) => {
    setFormData({
      ...formData,
      [event.target.name]: event.target.value,
    });
  };

  // Transmit data to the server's API for sending email.
  const handleSubmit = async (event) => {
    event.preventDefault(); // Prevent default form submission behavior

    const requestOptions = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(formData),
    };
    try {
      await fetch("http://localhost:5787/addItem", requestOptions);
    } catch (error) {
      console.error("Error fetching data:", error);
    }

    await getItems();
  };

  const getItems = async () => {
    try {
      var response = await fetch("http://localhost:5787/listItems", {
        method: "GET", // Use POST for sending data
        headers: { "Content-Type": "application/json" }, // Set the content type
      });
    } catch (error) {
      console.error("Error fetching data:", error);
    }

    const data = await response.json();
    setItems(data);
  }
  getItems();

  return (
    <div>
      <form onSubmit={handleSubmit}>
        <input
          placeholder="UPC"
          type="text"
          name="upc"
          value={formData.upc}
          onChange={handleChange}
        />
        <div className="row">
          <button type="submit" className="button shadow">
            Submit
          </button>
        </div>
      </form >
      <button onClick={getItems}>Get Items List</button>
      <table>
        <thead>
          <tr>
            <th>UPC</th>
            <th>Product Name</th>
            <th>Image</th>
            <th>Count</th>
            {/* Add more columns as needed */}
          </tr>
        </thead>
        <tbody>
          {items.map((item, index) => (
            <tr key={index}>
              <td>{item.upc}</td>
              <td>{item.name}</td>
              <td>
                <img src={item.image} />
              </td>
              <td>{item.count}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
export default App
