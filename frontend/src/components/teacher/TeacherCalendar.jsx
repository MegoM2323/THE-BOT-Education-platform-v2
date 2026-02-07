import { useState, useEffect } from "react";
import Calendar from "../admin/Calendar.jsx";
import "./TeacherCalendar.css";

export const TeacherCalendar = () => {
  return (
    <div className="teacher-calendar">
      <Calendar />
    </div>
  );
};

export default TeacherCalendar;
