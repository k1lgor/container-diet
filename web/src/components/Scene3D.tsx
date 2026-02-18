import { useRef, useMemo } from "react";
import { Canvas, useFrame } from "@react-three/fiber";
import { Float, PerspectiveCamera } from "@react-three/drei";
import * as THREE from "three";

// Floating Docker Container Box
interface ContainerBoxProps {
  position: [number, number, number];
  rotation?: [number, number, number];
  scale?: number;
  color?: string;
  speed?: number;
  delay?: number;
}

function ContainerBox({
  position,
  rotation = [0, 0, 0],
  scale = 1,
  color = "#2496ED",
  speed = 1,
  delay = 0,
}: ContainerBoxProps) {
  const meshRef = useRef<THREE.Mesh>(null!);
  const edgesRef = useRef<THREE.LineSegments>(null!);

  useFrame((state) => {
    const time = state.clock.getElapsedTime() + delay;
    if (meshRef.current) {
      meshRef.current.rotation.x =
        rotation[0] + Math.sin(time * speed * 0.3) * 0.15;
      meshRef.current.rotation.y = rotation[1] + time * speed * 0.2;
      meshRef.current.rotation.z =
        rotation[2] + Math.cos(time * speed * 0.2) * 0.08;
      meshRef.current.position.y =
        position[1] + Math.sin(time * speed * 0.5) * 0.3;
    }
    if (edgesRef.current) {
      edgesRef.current.rotation.x = meshRef.current?.rotation.x || 0;
      edgesRef.current.rotation.y = meshRef.current?.rotation.y || 0;
      edgesRef.current.rotation.z = meshRef.current?.rotation.z || 0;
      edgesRef.current.position.y = meshRef.current?.position.y || position[1];
    }
  });

  const geometry = useMemo(() => new THREE.BoxGeometry(1, 0.6, 1.6), []);
  const edgesGeometry = useMemo(
    () => new THREE.EdgesGeometry(geometry),
    [geometry],
  );

  return (
    <group>
      <mesh ref={meshRef} position={position} scale={scale} geometry={geometry}>
        <meshStandardMaterial
          color={color}
          transparent
          opacity={0.08}
          roughness={0.1}
          metalness={0.9}
          side={THREE.DoubleSide}
        />
      </mesh>
      <lineSegments
        ref={edgesRef}
        position={position}
        scale={scale}
        geometry={edgesGeometry}
      >
        <lineBasicMaterial
          color={color}
          transparent
          opacity={0.6}
          linewidth={1}
        />
      </lineSegments>
    </group>
  );
}

// Stacked Container Layers (the Docker "layers" metaphor)
function ContainerStack({ position }: { position: [number, number, number] }) {
  const groupRef = useRef<THREE.Group>(null!);

  useFrame((state) => {
    const time = state.clock.getElapsedTime();
    if (groupRef.current) {
      groupRef.current.rotation.y = time * 0.15;
      groupRef.current.position.y = position[1] + Math.sin(time * 0.4) * 0.4;
    }
  });

  const layers = [
    { y: 0, color: "#2496ED", opacity: 0.7 },
    { y: 0.35, color: "#086DD7", opacity: 0.6 },
    { y: 0.7, color: "#00BBFF", opacity: 0.5 },
    { y: 1.05, color: "#75AADB", opacity: 0.4 },
  ];

  return (
    <group ref={groupRef} position={position}>
      {layers.map((layer, i) => (
        <group key={i}>
          <mesh position={[0, layer.y, 0]}>
            <boxGeometry args={[1.8 - i * 0.1, 0.28, 1.8 - i * 0.1]} />
            <meshStandardMaterial
              color={layer.color}
              transparent
              opacity={0.06}
              roughness={0.2}
              metalness={0.8}
            />
          </mesh>
          <lineSegments position={[0, layer.y, 0]}>
            <edgesGeometry
              args={[new THREE.BoxGeometry(1.8 - i * 0.1, 0.28, 1.8 - i * 0.1)]}
            />
            <lineBasicMaterial
              color={layer.color}
              transparent
              opacity={layer.opacity}
            />
          </lineSegments>
        </group>
      ))}
    </group>
  );
}

// Data Packet Particles (network traffic visualization)
function DataPackets({ count = 120 }) {
  const pointsRef = useRef<THREE.Points>(null!);

  const { positions, velocities } = useMemo(() => {
    const positions = new Float32Array(count * 3);
    const velocities = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
      positions[i * 3] = (Math.random() - 0.5) * 20;
      positions[i * 3 + 1] = (Math.random() - 0.5) * 20;
      positions[i * 3 + 2] = (Math.random() - 0.5) * 15;
      velocities[i * 3] = (Math.random() - 0.5) * 0.02;
      velocities[i * 3 + 1] = (Math.random() - 0.5) * 0.02;
      velocities[i * 3 + 2] = (Math.random() - 0.5) * 0.01;
    }
    return { positions, velocities };
  }, [count]);

  useFrame(() => {
    if (!pointsRef.current) return;
    const pos = pointsRef.current.geometry.attributes.position
      .array as Float32Array;
    for (let i = 0; i < count; i++) {
      pos[i * 3] += velocities[i * 3];
      pos[i * 3 + 1] += velocities[i * 3 + 1];
      pos[i * 3 + 2] += velocities[i * 3 + 2];

      // Wrap around
      if (Math.abs(pos[i * 3]) > 10) velocities[i * 3] *= -1;
      if (Math.abs(pos[i * 3 + 1]) > 10) velocities[i * 3 + 1] *= -1;
      if (Math.abs(pos[i * 3 + 2]) > 7.5) velocities[i * 3 + 2] *= -1;
    }
    pointsRef.current.geometry.attributes.position.needsUpdate = true;
  });

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions}
          itemSize={3}
          args={[positions, 3]}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.05}
        color="#2496ED"
        transparent
        opacity={0.7}
        sizeAttenuation
      />
    </points>
  );
}

// Network Grid Lines (Docker network visualization)
function NetworkGrid() {
  const gridRef = useRef<THREE.Group>(null!);

  useFrame((state) => {
    const time = state.clock.getElapsedTime();
    if (gridRef.current) {
      gridRef.current.rotation.x = Math.sin(time * 0.05) * 0.1;
      gridRef.current.rotation.y = time * 0.03;
    }
  });

  const lineObjects = useMemo(() => {
    const nodes: THREE.Vector3[] = [];

    // Create network nodes
    for (let i = 0; i < 12; i++) {
      nodes.push(
        new THREE.Vector3(
          (Math.random() - 0.5) * 14,
          (Math.random() - 0.5) * 10,
          (Math.random() - 0.5) * 8,
        ),
      );
    }

    const material = new THREE.LineBasicMaterial({
      color: "#2496ED",
      transparent: true,
      opacity: 0.08,
    });

    const objects: THREE.Line[] = [];
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        if (nodes[i].distanceTo(nodes[j]) < 7) {
          const geo = new THREE.BufferGeometry().setFromPoints([
            nodes[i],
            nodes[j],
          ]);
          objects.push(new THREE.Line(geo, material));
        }
      }
    }
    return objects;
  }, []);

  return (
    <group ref={gridRef}>
      {lineObjects.map((lineObj, i) => (
        <primitive key={i} object={lineObj} />
      ))}
    </group>
  );
}

// Orbiting Ring (Docker hub / registry ring)
function OrbitalRing({ radius = 4, speed = 0.1, color = "#2496ED" }) {
  const ringRef = useRef<THREE.Mesh>(null!);

  useFrame((state) => {
    const time = state.clock.getElapsedTime();
    if (ringRef.current) {
      ringRef.current.rotation.x = Math.PI / 3 + Math.sin(time * 0.2) * 0.1;
      ringRef.current.rotation.y = time * speed;
      ringRef.current.rotation.z = Math.cos(time * 0.15) * 0.05;
    }
  });

  return (
    <mesh ref={ringRef}>
      <torusGeometry args={[radius, 0.015, 8, 80]} />
      <meshStandardMaterial
        color={color}
        transparent
        opacity={0.3}
        emissive={color}
        emissiveIntensity={0.5}
      />
    </mesh>
  );
}

// Glowing Core Sphere (Docker whale body)
function GlowCore() {
  const meshRef = useRef<THREE.Mesh>(null!);

  useFrame((state) => {
    const time = state.clock.getElapsedTime();
    if (meshRef.current) {
      meshRef.current.rotation.y = time * 0.1;
      (
        meshRef.current.material as THREE.MeshStandardMaterial
      ).emissiveIntensity = 0.3 + Math.sin(time * 1.5) * 0.15;
    }
  });

  return (
    <Float speed={0.8} rotationIntensity={0.2} floatIntensity={0.5}>
      <mesh ref={meshRef} position={[0, 0, -6]}>
        <sphereGeometry args={[1.8, 32, 32]} />
        <meshStandardMaterial
          color="#2496ED"
          transparent
          opacity={0.05}
          roughness={0.1}
          metalness={1}
          emissive="#2496ED"
          emissiveIntensity={0.3}
        />
      </mesh>
    </Float>
  );
}

export function Scene3D() {
  return (
    <div
      className="scene-container"
      style={{
        position: "absolute",
        top: 0,
        left: 0,
        width: "100%",
        height: "100%",
        zIndex: 0,
        pointerEvents: "none",
        background:
          "radial-gradient(ellipse at 30% 20%, rgba(36,150,237,0.12) 0%, transparent 50%), radial-gradient(ellipse at 70% 80%, rgba(8,109,215,0.08) 0%, transparent 50%), linear-gradient(180deg, #050d1a 0%, #0a1628 50%, #050d1a 100%)",
      }}
    >
      <Canvas dpr={[1, 1.5]} gl={{ antialias: true, alpha: true }}>
        <PerspectiveCamera makeDefault position={[0, 0, 8]} fov={60} />

        {/* Lighting - Docker Blue atmosphere */}
        <ambientLight intensity={0.2} color="#0a1628" />
        <pointLight position={[0, 0, 5]} intensity={2} color="#2496ED" />
        <pointLight position={[-8, 4, -2]} intensity={1.5} color="#086DD7" />
        <pointLight position={[8, -4, -2]} intensity={1} color="#00BBFF" />
        <spotLight
          position={[0, 10, 5]}
          angle={0.4}
          penumbra={1}
          intensity={3}
          color="#2496ED"
          castShadow={false}
        />

        {/* Core glow */}
        <GlowCore />

        {/* Orbital rings */}
        <OrbitalRing radius={4.5} speed={0.08} color="#2496ED" />
        <OrbitalRing radius={3.2} speed={-0.12} color="#00BBFF" />
        <OrbitalRing radius={5.8} speed={0.05} color="#086DD7" />

        {/* Floating container boxes */}
        <Float speed={1.2} rotationIntensity={0.3} floatIntensity={0.8}>
          <ContainerBox
            position={[3.5, 1.5, -3]}
            scale={0.9}
            color="#2496ED"
            speed={0.8}
            delay={0}
          />
        </Float>
        <Float speed={0.9} rotationIntensity={0.4} floatIntensity={1}>
          <ContainerBox
            position={[-3.8, -1.2, -4]}
            scale={1.1}
            color="#086DD7"
            speed={0.6}
            delay={1.5}
          />
        </Float>
        <Float speed={1.5} rotationIntensity={0.2} floatIntensity={0.6}>
          <ContainerBox
            position={[2, -2.5, -5]}
            scale={0.7}
            color="#00BBFF"
            speed={1.2}
            delay={3}
          />
        </Float>
        <Float speed={0.7} rotationIntensity={0.5} floatIntensity={1.2}>
          <ContainerBox
            position={[-2.5, 2.5, -6]}
            scale={1.3}
            color="#2496ED"
            speed={0.5}
            delay={0.8}
          />
        </Float>
        <Float speed={1.1} rotationIntensity={0.3} floatIntensity={0.9}>
          <ContainerBox
            position={[4.5, -0.5, -7]}
            scale={0.8}
            color="#75AADB"
            speed={0.9}
            delay={2.2}
          />
        </Float>

        {/* Container stacks */}
        <ContainerStack position={[-5, 0, -5]} />
        <ContainerStack position={[5, -1, -6]} />

        {/* Network visualization */}
        <NetworkGrid />

        {/* Data packets */}
        <DataPackets count={150} />
      </Canvas>
    </div>
  );
}
