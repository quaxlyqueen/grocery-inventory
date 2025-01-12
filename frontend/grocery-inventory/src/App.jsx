import React, { useState } from "react";

const App = () => {
  const [formData, setFormData] = useState({
    upc: "",
    exp_date: "",
    storage_id: 1,
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

    try {
      await fetch("http://localhost:5787/addItem", {
        method: "POST", // Use POST for sending data
        headers: { "Content-Type": "application/json" }, // Set the content type
        body: JSON.stringify(formData), // Convert the JS object to JSON string
      });
    } catch (error) {
      console.error("Error fetching data:", error);
    }
  };

  const getGroceries = async () => {
    try {
      var response = await fetch("http://localhost:5787/listGroceries", {
        method: "POST", // Use POST for sending data
        headers: { "Content-Type": "application/json" }, // Set the content type
        body: JSON.stringify(""), // Convert the JS object to JSON string
      });
    } catch (error) {
      console.error("Error fetching data:", error);
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        placeholder="UPC"
        type="text"
        name="upc"
        value={formData.upc}
        onChange={handleChange}
      />
      <input
        placeholder="exp_date"
        type="text"
        name="exp_date"
        value={formData.exp_date}
        onChange={handleChange}
      />
      <input
        placeholder="storage_id"
        type="text"
        name="storage_id"
        value={formData.storage_id}
        onChange={handleChange}
      />

      <div className="row">
        <button type="submit" className="button shadow">
          Submit
        </button>
      </div>
    </form >
  );
}
export default App
