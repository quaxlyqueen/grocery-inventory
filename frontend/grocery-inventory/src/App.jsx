import React, { useState } from "react";

const App = () => {
  const [formData, setFormData] = useState({
    upc: ""
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
    console.log(requestOptions);
    try {
      await fetch("http://localhost:5787/addItem", requestOptions);
    } catch (error) {
      console.error("Error fetching data:", error);
    }
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

    console.log(response.body);
  }

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
    </div>
  );
}
export default App
