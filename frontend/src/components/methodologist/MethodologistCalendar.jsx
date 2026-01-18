import { useState, useEffect } from "react";
import Calendar from "../admin/Calendar.jsx";
import "./MethodologistCalendar.css";

export const MethodologistCalendar = () => {
  return (
    <div className="methodologist-calendar">
      <Calendar />
    </div>
  );
};

export default MethodologistCalendar;
