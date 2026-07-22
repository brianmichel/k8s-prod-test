<script setup lang="ts">
import { computed, type Component } from 'vue'
import {
  ArrowPathIcon,
  ChartBarIcon,
  CircleStackIcon,
  CubeIcon,
  DocumentTextIcon,
  FolderIcon,
  LockClosedIcon,
  QuestionMarkCircleIcon,
  ShareIcon,
  Square3Stack3DIcon,
  Squares2X2Icon,
  UserCircleIcon,
} from '@heroicons/vue/24/outline'

const props=defineProps<{kind:string}>()
const icons:Readonly<Record<string,Component>>={
  Application:Squares2X2Icon,
  Deployment:ArrowPathIcon,
  Rollout:ShareIcon,
  ReplicaSet:Square3Stack3DIcon,
  Pod:CubeIcon,
  Service:CircleStackIcon,
  ConfigMap:DocumentTextIcon,
  Secret:LockClosedIcon,
  ServiceAccount:UserCircleIcon,
  Namespace:FolderIcon,
  AnalysisRun:ChartBarIcon,
}
const icon=computed(()=>icons[props.kind]??QuestionMarkCircleIcon)
</script>

<template><component :is="icon" aria-hidden="true" /></template>
