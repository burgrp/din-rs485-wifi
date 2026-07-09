from register_definition import Register

SinotimerEnergyMeter3P = (
    Register(0x0000, 'voltage.1', 'Voltage L1', 'V'),
    Register(0x0002, 'voltage.2', 'Voltage L2', 'V'),
    Register(0x0004, 'voltage.3', 'Voltage L3', 'V'),
    Register(0x0008, 'current.1', 'Current L1', 'A'),
    Register(0x000A, 'current.2', 'Current L2', 'A'),
    Register(0x000C, 'current.3', 'Current L3', 'A'),
    Register(0x0010, 'power.active.total', 'Total active power', 'W'),
    Register(0x0012, 'power.active.1', 'Active power L1', 'W'),
    Register(0x0014, 'power.active.2', 'Active power L2', 'W'),
    Register(0x0016, 'power.active.3', 'Active power L3', 'W'),
    Register(0x0018, 'power.reactive.total', 'Total reactive power', 'VA'),
    Register(0x001A, 'power.reactive.1', 'Reactive power L1', 'VA'),
    Register(0x001C, 'power.reactive.2', 'Reactive power L2', 'VA'),
    Register(0x001E, 'power.reactive.3', 'Reactive power L3', 'VA'),
    Register(0x002A, 'power.factor.1', 'Power factor L1', None),
    Register(0x002C, 'power.factor.2', 'Power factor L2', None),
    Register(0x002E, 'power.factor.3', 'Power factor L3', None),
    Register(0x0036, 'frequency', 'Frequency', 'Hz'),
    Register(0x0100, 'energy.active', 'Total active energy', 'kWh'),
    Register(0x0400, 'energy.reactive', 'Total reactive energy', 'kVAh'),
)
