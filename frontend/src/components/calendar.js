// Calendario estático, 100% frontend (sin backend). Mes navegable.

export function calendar(firstDayOfWeek = "monday") {
  const now = new Date();
  return {
    year: now.getFullYear(),
    month: now.getMonth(), // 0..11
    firstDay: firstDayOfWeek === "sunday" ? 0 : 1,

    get monthLabel() {
      return new Date(this.year, this.month, 1).toLocaleDateString("es", {
        month: "long",
        year: "numeric",
      });
    },
    get weekdays() {
      const base = ["Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"];
      return [...base.slice(this.firstDay), ...base.slice(0, this.firstDay)];
    },
    get days() {
      const first = new Date(this.year, this.month, 1);
      const offset = (first.getDay() - this.firstDay + 7) % 7;
      const daysInMonth = new Date(this.year, this.month + 1, 0).getDate();
      const cells = [];

      // Días del mes anterior (relleno)
      const prevDays = new Date(this.year, this.month, 0).getDate();
      for (let i = offset - 1; i >= 0; i--) {
        cells.push({ day: prevDays - i, current: false, today: false });
      }
      // Días del mes actual
      const today = new Date();
      for (let d = 1; d <= daysInMonth; d++) {
        const isToday =
          d === today.getDate() && this.month === today.getMonth() && this.year === today.getFullYear();
        cells.push({ day: d, current: true, today: isToday });
      }
      // Relleno final hasta completar semanas
      while (cells.length % 7 !== 0) {
        cells.push({ day: cells.length, current: false, today: false });
      }
      return cells;
    },
    prev() {
      if (this.month === 0) {
        this.month = 11;
        this.year--;
      } else this.month--;
    },
    next() {
      if (this.month === 11) {
        this.month = 0;
        this.year++;
      } else this.month++;
    },
  };
}
