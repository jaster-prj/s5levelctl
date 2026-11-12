"""Minimaler Dummy der für S5LevelCtl relevanten Griffin-Konfiguration."""

class ProbeMode:
    DETAILED = "DETAILED"
    NEVER = "NEVER"


class AlignZAxisProcedure:
    def __init__(self):
        # Diese Zeile entspricht der von Griffin ausgewerteten Einstellung.
        self.__probing_mode = ProbeMode.DETAILED
