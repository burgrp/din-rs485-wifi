from sinotimer_energy_meter_3p import *

devices=(
    {
        "name": "em.grid",
        "registers": SinotimerEnergyMeter3P,
        "address": 1,
        "baud": 9600
    },
    {
        "name": "em.house",
        "registers": SinotimerEnergyMeter3P,
        "address": 2,
        "baud": 9600
    },
)
debug=True
emu=True
